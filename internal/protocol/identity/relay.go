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

func (r *Relay) HydrateIdentity(uuid [16]byte, pubKey [32]byte) {
	r.UUID = uuid
	r.PubKey = pubKey
}
