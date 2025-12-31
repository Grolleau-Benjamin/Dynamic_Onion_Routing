package client

import (
	"fmt"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
)

type GroupCryptoContext struct {
	Group identity.RelayGroup

	CipherKey [32]byte
	EPK       [32]byte
	ESK       [32]byte
}

func (g GroupCryptoContext) String() string {
	return fmt.Sprintf(
		"{\n\t\tgroup=%s cipher=%x epk=%x}",
		g.Group,
		g.CipherKey[:4],
		g.EPK[:4],
	)
}

type Message struct {
	Dest identity.Endpoint
	Path []GroupCryptoContext

	Payload []byte
}

func (m Message) String() string {
	return fmt.Sprintf(
		"Message{\n\tdest=%s \n\tpath=%v \n\tpayload_len=%d\n}",
		m.Dest,
		m.Path,
		len(m.Payload),
	)
}
