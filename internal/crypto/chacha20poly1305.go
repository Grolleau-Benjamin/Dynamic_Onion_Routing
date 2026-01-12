package crypto

import "golang.org/x/crypto/chacha20poly1305"

const (
	Poly1305TagSize = 16
)

func ChachaEncrypt(
	key [32]byte,
	nonce [12]byte,
	plaintext []byte,
	aad []byte,
) ([]byte, error) {
	aead, err := chacha20poly1305.New(key[:])
	if err != nil {
		return nil, err
	}

	return aead.Seal(nil, nonce[:], plaintext, aad), nil
}

func ChachaDecrypt(
	key [32]byte,
	nonce [12]byte,
	ciphertext []byte,
	aad []byte,
) ([]byte, error) {
	aead, err := chacha20poly1305.New(key[:])
	if err != nil {
		return nil, err
	}

	return aead.Open(nil, nonce[:], ciphertext, aad)
}
