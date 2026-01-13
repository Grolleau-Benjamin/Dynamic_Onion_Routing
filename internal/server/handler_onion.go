package server

import (
	"bytes"
	"net"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/crypto"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/onion"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/transport"
	"golang.org/x/crypto/curve25519"
)

func handleOnionPacket(
	p packet.Packet,
	conn net.Conn,
	s *Server,
) {
	logger.Debugf("[%s] Onion packet received", conn.RemoteAddr())

	onionPkt, ok := p.(*packet.OnionPacket)
	if !ok {
		logger.Warnf("[%s] Failed to cast packet to OnionPacket", conn.RemoteAddr())
		return
	}

	layer := &onion.OnionLayer{}
	err := layer.Parse(onionPkt.Data[:])
	if err != nil {
		logger.Warnf("[%s] Failed to parse onion layer: %v", conn.RemoteAddr(), err)
		return
	}

	logger.Debugf("[%s] Onion Layer informations: {\n\tepk: %x\n\twks: [%d]\n\tflags: %d\n\tnonce: %x\n\tct: %x...\n}",
		conn.RemoteAddr(),
		layer.EPK,
		len(layer.WrappedKeys),
		layer.Flags,
		layer.PayloadNonce,
		layer.CipherText[:3],
	)

	sharedSecret, err := curve25519.X25519(s.Pi.PrivKey[:], layer.EPK[:])
	if err != nil {
		logger.Warnf("[%s] Failed to generate shared secret: %v", conn.RemoteAddr(), err)
		return
	}

	wrappingKeySlice, err := crypto.HKDFSha256(
		sharedSecret,
		onion.HKDFSaltWrappedKey,
		onion.HKDFInfoWrappedKey,
	)
	if err != nil {
		logger.Warnf("[%s] Failed to generate wrappingKeySlice: %v", conn.RemoteAddr(), err)
		return
	}

	var wrappingKey [32]byte
	var sessionKey [32]byte
	found := false
	copy(wrappingKey[:], wrappingKeySlice)

	for _, wk := range layer.WrappedKeys {
		res, err := crypto.ChachaDecrypt(
			wrappingKey,
			wk.Nonce,
			wk.CipherText[:],
			[]byte("DORv1:WrappedKey"),
		)
		if err != nil {
			// Not the right WK
			continue
		}

		if bytes.Equal(res[0:16], s.Pi.UUID[:]) {
			copy(sessionKey[:], res[16:48])
			found = true
			break
		}
	}

	if !found {
		logger.Warnf("[%s] No matching wrapped key found (not in this route)", conn.RemoteAddr())
		return
	}

	if err = layer.TrimCipherText(sessionKey); err != nil {
		logger.Warnf("[%s] failed to trim CipherText: %v", conn.RemoteAddr(), err)
	}

	logger.Debugf("[%s] Real cipher text length: %d bytes", conn.RemoteAddr(), len(layer.CipherText))

	headerBytes, err := layer.HeaderBytes()
	if err != nil {
		logger.Warnf("[%s] Failed to get header bytes: %v", conn.RemoteAddr(), err)
		return
	}

	plaintext, err := crypto.ChachaDecrypt(
		sessionKey,
		layer.PayloadNonce,
		layer.CipherText,
		headerBytes,
	)
	if err != nil {
		logger.Warnf("[%s] failed to decrypt layer.Ciphertext: %v", conn.RemoteAddr(), err)
	}

	var olc onion.OnionLayerCiphered
	if err = olc.Parse(plaintext); err != nil {
		logger.Warnf("[%s] failed to parse plaintext: %v", conn.RemoteAddr(), err)
	}

	logger.Debugf("[%s] OnionLayerCiphered parsed: {\n\tLastServer: %v\n\tNextHops: %d\n\tUtilPayloadLength: %d\n\tPayload: %d bytes\n}",
		conn.RemoteAddr(),
		olc.LastServer,
		len(olc.NextHops),
		olc.UtilPayloadLength,
		len(olc.Payload),
	)
	for i, nh := range olc.NextHops {
		logger.Debugf("[%s] NextHop[%d]: %s:%d",
			conn.RemoteAddr(), i, nh.IP.String(), nh.Port)
	}

	if olc.LastServer {
		// TODO: handler this later
		logger.Infof("[%s] Final destination reached! Processing payload (%d bytes)...", conn.RemoteAddr(), len(olc.Payload))
		return
	}

	if len(olc.NextHops) == 0 {
		logger.Warnf("[%s] Relay node but no next hop defined!", conn.RemoteAddr())
		return
	}

	nextLayer := &onion.OnionLayer{}
	err = nextLayer.Parse(olc.Payload)
	if err != nil {
		logger.Warnf("[%s] Decrypted payload is not a valid OnionLayer: %v", conn.RemoteAddr(), err)
		return
	}

	nextHopBytes, err := nextLayer.BytesPadded()
	if err != nil {
		logger.Warnf("[%s] Failed to pad next layer: %v", conn.RemoteAddr(), err)
		return
	}

	var outPkt packet.OnionPacket
	copy(outPkt.Data[:], nextHopBytes)
	trans := transport.NewTransport()

	for _, nh := range olc.NextHops {
		if err := trans.Send(nh, &outPkt); err != nil {
			logger.Warnf("[%s] Failed to relay packet to %s: %v", conn.RemoteAddr(), nh.String(), err)
			continue
		}
		logger.Debugf("[%s] Packet successfully relayed to %s", conn.RemoteAddr(), nh.String())
		return
	}
}
