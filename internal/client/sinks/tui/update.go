package tui

import (
	"fmt"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.state.Client.Close()
			return m, tea.Quit
		}
	case client.Event:
		m.state.Logs = append(m.state.Logs, fmt.Sprintf("%v", msg.Payload))
		return m, waitForEvent(m.state.Client.Events())
	}
	return m, nil
}

func waitForEvent(ch <-chan client.Event) tea.Cmd {
	return func() tea.Msg {
		if event, ok := <-ch; ok {
			return event
		}
		return nil
	}
}
