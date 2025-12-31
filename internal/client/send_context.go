package client

import (
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
)

type SendContext struct {
	Message Message
}

func BuildSendContext(destStr, pathStr string, payload []byte) (SendContext, error) {
	dest, err := identity.ParseEpFromString(destStr)
	if err != nil {
		return SendContext{}, err
	}

	groups, err := identity.ParseRelayPath(pathStr)
	if err != nil {
		return SendContext{}, err
	}

	path := make([]GroupCryptoContext, 0, len(groups))
	for _, g := range groups {
		path = append(path, GroupCryptoContext{
			Group: g,
		})
	}

	msg := Message{
		Dest:    dest,
		Path:    path,
		Payload: payload,
	}

	return SendContext{Message: msg}, nil
}
