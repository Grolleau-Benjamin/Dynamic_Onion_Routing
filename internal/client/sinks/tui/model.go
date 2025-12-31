package tui

import (
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	state State
}

func New(c *client.Client, ic client.InputConfig) *Model {
	return &Model{
		state: State{
			Client: c,
			Config: ic,
			Logs:   make([]string, 0),
		},
	}
}

func (m *Model) Start() error {
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

func (m *Model) Init() tea.Cmd {
	m.state.Client.Simulate()
	return waitForEvent(m.state.Client.Events())
}
