package server

import (
	"bytes"
	"fmt"
	"net"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/crypto"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/onion"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/transport"
	"golang.org/x/crypto/curve25519"
)

func handleOnionPacket(p packet.Packet, conn net.Conn, s *Server) {
	logger.Debugf("[%s] Onion packet received", conn.RemoteAddr())

	onionPkt, err := assertOnionPacketType(
		p,
		conn,
	)
	if err != nil {
		return
	}

	layer, err := parseInboundLayer(
		onionPkt,
		conn,
	)
	if err != nil {
		return
	}

	sessionKey, err := unwrapSessionKey(
		layer,
		s,
		conn,
	)
	if err != nil {
		return
	}

	olc, err := decryptNextLayer(
		layer,
		sessionKey,
		conn,
	)
	if err != nil {
		return
	}

	if olc.LastServer {
		handleFinalDestination(
			olc,
			conn,
		)
		return
	}

	relayToNextHops(
		olc,
		conn,
	)
}

func assertOnionPacketType(p packet.Packet, conn net.Conn) (*packet.OnionPacket, error) {
	onionPkt, ok := p.(*packet.OnionPacket)
	if !ok {
		logger.Warnf("[%s] Failed to cast packet to OnionPacket", conn.RemoteAddr())
		return nil, fmt.Errorf("invalid type, wanted OnionPacket %d, got %d",
			packet.TypeOnionPacket, p.Type(),
		)
	}
	return onionPkt, nil
}

func parseInboundLayer(pkt *packet.OnionPacket, conn net.Conn) (*onion.OnionLayer, error) {
	layer := &onion.OnionLayer{}
	if err := layer.Parse(pkt.Data[:]); err != nil {
		logger.Warnf("[%s] Failed to parse onion layer: %v", conn.RemoteAddr(), err)
		return nil, err
	}

	logger.Debugf("[%s] Onion Layer informations: {\n\tepk: %x\n\twrappingKeySlice: [%d]\n\tflags: %d\n\tnonce: %x\n\tct: %x...\n}",
		conn.RemoteAddr(),
		layer.EPK,
		len(layer.WrappedKeys),
		layer.Flags,
		layer.PayloadNonce,
		layer.CipherText[:3],
	)
	return layer, nil
}

func unwrapSessionKey(layer *onion.OnionLayer, s *Server, conn net.Conn) ([32]byte, error) {
	sharedSecret, err := curve25519.X25519(s.Pi.PrivKey[:], layer.EPK[:])
	if err != nil {
		logger.Warnf("[%s] Failed to generate shared secret: %v", conn.RemoteAddr(), err)
		return [32]byte{}, err
	}

	wrappingKeySlice, err := crypto.HKDFSha256(
		sharedSecret,
		onion.HKDFSaltWrappedKey,
		onion.HKDFInfoWrappedKey,
	)
	if err != nil {
		logger.Warnf("[%s] Failed to generate wrappingKeySlice: %v", conn.RemoteAddr(), err)
		return [32]byte{}, err
	}

	var wrappingKey [32]byte
	copy(wrappingKey[:], wrappingKeySlice)

	for _, wk := range layer.WrappedKeys {
		res, err := crypto.ChachaDecrypt(
			wrappingKey,
			wk.Nonce,
			wk.CipherText[:],
			[]byte("DORv1:WrappedKey"),
		)
		if err != nil {
			continue
		}

		if bytes.Equal(res[:16], s.Pi.UUID[:]) {
			var sessionKey [32]byte
			copy(sessionKey[:], res[16:48])
			return sessionKey, nil
		}
	}

	logger.Warnf("[%s] No matching wrapped key found (not in this route)", conn.RemoteAddr())
	return [32]byte{}, fmt.Errorf("no matching wrapped key")
}

func decryptNextLayer(layer *onion.OnionLayer, sessionKey [32]byte, conn net.Conn) (*onion.OnionLayerCiphered, error) {
	if err := layer.TrimCipherText(sessionKey); err != nil {
		logger.Warnf("[%s] failed to trim CipherText: %v", conn.RemoteAddr(), err)
		return nil, err
	}

	header, err := layer.HeaderBytes()
	if err != nil {
		logger.Warnf("[%s] Failed to get header bytes: %v", conn.RemoteAddr(), err)
		return nil, err
	}

	plaintext, err := crypto.ChachaDecrypt(
		sessionKey,
		layer.PayloadNonce,
		layer.CipherText,
		header,
	)
	if err != nil {
		logger.Warnf("[%s] failed to decrypt layer.Ciphertext: %v", conn.RemoteAddr(), err)
		return nil, err
	}

	var olc onion.OnionLayerCiphered
	if err := olc.Parse(plaintext); err != nil {
		logger.Warnf("[%s] failed to parse plaintext: %v", conn.RemoteAddr(), err)
		return nil, err
	}

	logger.Debugf("[%s] OnionLayerCiphered parsed: {\n\tLastServer: %v\n\tNextHops: %d\n\tUtilPayloadLength: %d\n\tPayload: %d bytes\n}",
		conn.RemoteAddr(), olc.LastServer, len(olc.NextHops), olc.UtilPayloadLength, len(olc.Payload),
	)

	return &olc, nil
}

func handleFinalDestination(olc *onion.OnionLayerCiphered, conn net.Conn) {
	logger.Infof("[%s] Final destination reached! Processing payload (%d bytes)...",
		conn.RemoteAddr(), len(olc.Payload),
	)
}

func relayToNextHops(olc *onion.OnionLayerCiphered, conn net.Conn) {
	if len(olc.NextHops) == 0 {
		logger.Warnf("[%s] Relay node but no next hop defined!", conn.RemoteAddr())
		return
	}

	nextLayer := &onion.OnionLayer{}
	if err := nextLayer.Parse(olc.Payload); err != nil {
		logger.Warnf("[%s] Decrypted payload is not a valid OnionLayer: %v", conn.RemoteAddr(), err)
		return
	}

	bytes, err := nextLayer.BytesPadded()
	if err != nil {
		logger.Warnf("[%s] Failed to pad next layer: %v", conn.RemoteAddr(), err)
		return
	}

	var outPkt packet.OnionPacket
	copy(outPkt.Data[:], bytes)

	trans := transport.NewTransport()
	for _, nh := range olc.NextHops {
		if err := trans.Send(nh, &outPkt); err != nil {
			logger.Warnf("[%s] Failed to relay packet to %s: %v",
				conn.RemoteAddr(), nh.String(), err,
			)
			continue
		}
		logger.Debugf("[%s] Packet successfully relayed to %s",
			conn.RemoteAddr(), nh.String(),
		)
		return
	}
}
