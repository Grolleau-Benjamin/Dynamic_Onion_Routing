package identity

import (
	"fmt"
)

type Relay struct {
	Ep Endpoint

	UUID   [16]byte
	PubKey [32]byte
}

func (r Relay) String() string {
	return fmt.Sprintf(
		"{ep=%s uuid=%x pub=%x}",
		r.Ep,
		r.UUID[:4],
		r.PubKey[:4],
	)
}
