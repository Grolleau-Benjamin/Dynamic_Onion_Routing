package onion

import (
	"crypto/rand"
	"fmt"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/crypto"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
)

func BuildOnion(
	dest identity.Endpoint,
	path []identity.CryptoGroup,
	payload []byte,
) (*OnionLayer, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("path cannot be empty")
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
			return nil, fmt.Errorf("failed to generate payload nonce: %w", err)
		}

		cipherText, err := crypto.ChachaEncrypt(
			group.CipherKey,
			payloadNonce,
			cipheredBytes,
			[]byte("DORv1:OnionLayer"),
		)
		if err != nil {
			return nil, err
		}

		wrappedKeys, err := NewWrappedKeys(group)
		if err != nil {
			return nil, err
		}

		layer = &OnionLayer{
			EPK:          group.EPK,
			WrappedKeys:  wrappedKeys,
			Flags:        0x00,
			PayloadNonce: payloadNonce,
			CipherText:   cipherText,
		}

		currentPayload, err = layer.Bytes()
		if err != nil {
			return nil, err
		}
		nextHops = make([]identity.Endpoint, len(group.Group.Relays))
		for i, relay := range group.Group.Relays {
			nextHops[i] = relay.Ep
		}
		isLast = false
	}

	return layer, nil
}
