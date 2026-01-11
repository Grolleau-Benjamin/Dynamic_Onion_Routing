package crypto

import (
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/hkdf"
)

func HKDFSha256(
	secret []byte,
	salt []byte,
	info []byte,
) ([]byte, error) {
	h := hkdf.New(sha256.New, secret, salt, info)
	out := make([]byte, 32)

	if _, err := io.ReadFull(h, out); err != nil {
		return nil, err
	}

	return out, nil
}
