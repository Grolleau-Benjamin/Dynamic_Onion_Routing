package clienttui

import (
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	client   *client.Client
	events   []client.Event
	eventsCh <-chan client.Event
	done     bool

	// mode
	state State

	// Inputs values
	dest    identity.Endpoint
	path    []client.GroupCryptoContext
	payload string
}

func NewModel(c *client.Client, ch <-chan client.Event, input client.InputConfig) Model {
	m := Model{
		client:   c,
		eventsCh: ch,
		state:    NewState(),
		payload:  input.Payload,
	}

	if input.Dest != "" {
		if dest, err := identity.ParseEpFromString(input.Dest); err == nil {
			m.dest = dest
		}
	}

	if input.Path != "" {
		if groups, err := identity.ParseRelayPath(input.Path); err == nil {
			path := make([]client.GroupCryptoContext, 0, len(groups))
			for _, g := range groups {
				path = append(path, client.GroupCryptoContext{
					Group: g,
				})
			}
			m.path = path
		}
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return waitEvent(m.eventsCh)
}
