package identity

import (
	"fmt"
	"strings"
)

const (
	MaxJump int = 5 // todo: centralize with internal/protocol/onion/builder.go
	MaxNode int = 3 // todo: centralisze with MaxWrappedKey
)

type RelayGroup struct {
	Relays []Relay
}

func (g RelayGroup) String() string {
	var b strings.Builder
	b.WriteString("[")
	for i, r := range g.Relays {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(r.Ep.String())
	}
	b.WriteString("]")
	return b.String()
}

func ParseRelayPath(raw string) ([]RelayGroup, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty relay path")
	}

	groupStrs := strings.Split(raw, "|")
	if len(groupStrs) > MaxJump {
		return nil, fmt.Errorf("too many jump. Max: %d", MaxJump)
	}

	groups := make([]RelayGroup, 0, len(groupStrs))

	for gi, groupStr := range groupStrs {
		groupStr = strings.TrimSpace(groupStr)
		if groupStr == "" {
			return nil, fmt.Errorf("empty relay group at index %d", gi)
		}

		relayStrs := strings.Split(groupStr, ",")
		if len(relayStrs) > MaxNode {
			return nil, fmt.Errorf("too many node at group index %d. Max: %d", gi, MaxNode)
		}
		relays := make([]Relay, 0, len(relayStrs))

		for ri, relayStr := range relayStrs {
			relayStr = strings.TrimSpace(relayStr)
			if relayStr == "" {
				return nil, fmt.Errorf(
					"empty relay in group %d at index %d",
					gi, ri,
				)
			}

			ep, err := ParseEpFromString(relayStr)
			if err != nil {
				return nil, fmt.Errorf(
					"invalid relay %q in group %d: %w",
					relayStr, gi, err,
				)
			}

			relays = append(relays, Relay{
				Ep: ep,
			})
		}

		if len(relays) == 0 {
			return nil, fmt.Errorf("relay group %d has no relays", gi)
		}

		groups = append(groups, RelayGroup{
			Relays: relays,
		})
	}

	return groups, nil
}
