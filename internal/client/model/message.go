package model

import (
	"fmt"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
)

type Message struct {
	Dest    identity.Endpoint
	Path    []identity.CryptoGroup
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

func BuildFromInputConfig(ic client.InputConfig) (Message, error) {
	dest, err := identity.ParseEpFromString(ic.Dest)
	if err != nil {
		return Message{}, err
	}

	groups, err := identity.ParseRelayPath(ic.OnionPath)
	if err != nil {
		return Message{}, err
	}

	path := make([]identity.CryptoGroup, 0, len(groups))
	for _, g := range groups {
		path = append(path, identity.CryptoGroup{
			Group: g,
		})
	}

	return Message{
		Dest:    dest,
		Path:    path,
		Payload: []byte(ic.Payload),
	}, nil
}
