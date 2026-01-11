package onion

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/crypto"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"golang.org/x/crypto/curve25519"
)

type WrappedKey struct {
	Nonce      [12]byte
	CipherText [64]byte // [0:16] ruuid; [16:48] cipherKey; [48:64] ChaCha20 AuthTag
}

var (
	HKDFSaltWrappedKey = []byte("DORv1:WrappedKey")
	HKDFInfoWrappedKey = []byte("DORv1:RelayKeyEncryption")
)

func NewWrappedKeys(group *identity.CryptoGroup) ([MaxWrappedKey]WrappedKey, error) {
	var finalKeys [MaxWrappedKey]WrappedKey
	relays := group.Group.Relays

	if len(relays) > MaxWrappedKey {
		return finalKeys, fmt.Errorf("too many relays in group: %d (max %d)", len(relays), MaxWrappedKey)
	}

	for i, relay := range relays {
		sharedSecret, err := curve25519.X25519(group.ESK[:], relay.PubKey[:])
		if err != nil {
			return finalKeys, err
		}

		wrappingKeySlice, err := crypto.HKDFSha256(
			sharedSecret,
			HKDFSaltWrappedKey,
			HKDFInfoWrappedKey,
		)
		if err != nil {
			return finalKeys, err
		}

		var wrappingKey [32]byte
		copy(wrappingKey[:], wrappingKeySlice)

		plaintext := make([]byte, 48)
		copy(plaintext[0:16], relay.UUID[:])
		copy(plaintext[16:48], group.CipherKey[:])

		var wkNonce [12]byte
		if _, err = rand.Read(wkNonce[:]); err != nil {
			return finalKeys, fmt.Errorf("nonce gen failed: %w", err)
		}

		encryptedWK, err := crypto.ChachaEncrypt(
			wrappingKey,
			wkNonce,
			plaintext,
			[]byte("DORv1:WrappedKey"),
		)
		if err != nil {
			return finalKeys, err
		}

		if len(encryptedWK) != 64 {
			return finalKeys, fmt.Errorf("invalid wrapped key size: %d", len(encryptedWK))
		}

		var cipherTextArr [64]byte
		copy(cipherTextArr[:], encryptedWK)

		finalKeys[i] = WrappedKey{
			Nonce:      wkNonce,
			CipherText: cipherTextArr,
		}
	}

	for i := len(relays); i < MaxWrappedKey; i++ {
		var dummy WrappedKey
		if _, err := rand.Read(dummy.Nonce[:]); err != nil {
			return finalKeys, fmt.Errorf("dummy nonce gen failed: %w", err)
		}
		if _, err := rand.Read(dummy.CipherText[:]); err != nil {
			return finalKeys, fmt.Errorf("dummy ciphertext gen failed: %w", err)
		}
		finalKeys[i] = dummy
	}

	for i := len(finalKeys) - 1; i > 0; i-- {
		randIndexBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return finalKeys, fmt.Errorf("shuffle failed: %w", err)
		}
		j := int(randIndexBig.Int64())

		finalKeys[i], finalKeys[j] = finalKeys[j], finalKeys[i]
	}

	return finalKeys, nil
}
