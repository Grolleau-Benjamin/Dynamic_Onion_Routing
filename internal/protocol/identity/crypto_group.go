package identity

import (
	"fmt"
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
