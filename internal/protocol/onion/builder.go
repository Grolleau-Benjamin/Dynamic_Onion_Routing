package onion

import (
	"crypto/rand"
	"fmt"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/crypto"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
)

const (
	PacketSize = 4096
	MaxJump    = 5
)

func BuildOnion(
	dest identity.Endpoint,
	path []identity.CryptoGroup,
	payload []byte,
) (*OnionLayer, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("path cannot be empty")
	}
	if len(path) > MaxJump {
		return nil, fmt.Errorf("max jump value is %d", MaxJump)
	}

	overhead := computePathOverhead(path, dest)
	totalSize := overhead + len(payload)
	if totalSize > PacketSize {
		return nil, fmt.Errorf("payload too large: %d bytes (max allowed with this path: %d)", len(payload), PacketSize-overhead)
	}

	currentPayload := payload
	nextHops := []identity.Endpoint{dest}
	isLast := true

	var layer *OnionLayer

	for i := len(path) - 1; i >= 0; i-- {
		group := &path[i]
		ciphered := OnionLayerCiphered{
			LastServer:        isLast,
			NextHops:          nextHops,
			UtilPayloadLength: uint16(len(currentPayload)),
			Payload:           currentPayload,
		}

		cipheredBytes, err := ciphered.Bytes()
		if err != nil {
			return nil, err
		}

		var payloadNonce [12]byte
		if _, err = rand.Read(payloadNonce[:]); err != nil {
			return nil, fmt.Errorf("failed to generate payload nonce: %v", err)
		}

		wrappedKeys, err := NewWrappedKeys(group)
		if err != nil {
			return nil, err
		}

		expectedCipherLen := len(cipheredBytes) + crypto.Poly1305TagSize
		if expectedCipherLen > 0xFFFF {
			return nil, fmt.Errorf("ciphertext too large: %d", expectedCipherLen)
		}

		mask16, err := CipherTextLenMask16(group.CipherKey, payloadNonce)
		if err != nil {
			return nil, err
		}
		cipherLenXor := uint16(expectedCipherLen) ^ mask16

		layer = &OnionLayer{
			EPK:              group.EPK,
			WrappedKeys:      wrappedKeys,
			Flags:            0x00,
			PayloadNonce:     payloadNonce,
			CipherTextLenXor: cipherLenXor,
			CipherText:       nil,
		}

		headerBytes, err := layer.HeaderBytes()
		if err != nil {
			return nil, err
		}

		cipherText, err := crypto.ChachaEncrypt(
			group.CipherKey,
			payloadNonce,
			cipheredBytes,
			headerBytes,
		)
		if err != nil {
			return nil, err
		}
		if len(cipherText) != expectedCipherLen {
			return nil, fmt.Errorf("unexpected ciphertext size: got %d, want %d", len(cipherText), expectedCipherLen)
		}

		layer.CipherText = cipherText

		nextHops = make([]identity.Endpoint, len(group.Group.Relays))
		for i, relay := range group.Group.Relays {
			nextHops[i] = relay.Ep
		}
		isLast = false
		currentPayload, err = layer.Bytes()
		if err != nil {
			return nil, err
		}
	}

	return layer, nil
}

func computePathOverhead(path []identity.CryptoGroup, dest identity.Endpoint) int {
	overhead := 0

	overhead += InnerMetadataFixedSize + dest.BytesLen()

	for _, group := range path {
		overhead += FixedHeaderSize
		overhead += crypto.Poly1305TagSize

		currentStepOverhead := InnerMetadataFixedSize
		for _, ep := range group.Group.Relays {
			currentStepOverhead += ep.Ep.BytesLen()
		}
		overhead += currentStepOverhead
	}

	return overhead
}
