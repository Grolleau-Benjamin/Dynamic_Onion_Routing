package identity

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

type CryptoGroup struct {
	Group RelayGroup

	CipherKey [32]byte
	EPK       [32]byte
	ESK       [32]byte
}

func (g CryptoGroup) String() string {
	return fmt.Sprintf(
		"{group=%s cipher=%x epk=%x}",
		g.Group,
		g.CipherKey[:4],
		g.EPK[:4],
	)
}

func (g *CryptoGroup) GenerateCryptoMaterial() error {
	if _, err := rand.Read(g.CipherKey[:]); err != nil {
		return fmt.Errorf("cipher key generation failed: %w", err)
	}

	if _, err := rand.Read(g.ESK[:]); err != nil {
		return fmt.Errorf("esk generation failed: %w", err)
	}

	g.ESK[0] &= 248
	g.ESK[31] &= 127
	g.ESK[31] |= 64

	epk, err := curve25519.X25519(g.ESK[:], curve25519.Basepoint)
	if err != nil {
		return fmt.Errorf("epk derivation failed: %w", err)
	}
	copy(g.EPK[:], epk)

	return nil
}
