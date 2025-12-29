package clienttui

import (
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	client   *client.Client
	events   []client.Event
	eventsCh <-chan client.Event
	done     bool
}

func NewModel(c *client.Client, ch <-chan client.Event) Model {
	return Model{
		client:   c,
		eventsCh: ch,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		startClient(m.client),
		waitEvent(m.eventsCh),
	)
}
