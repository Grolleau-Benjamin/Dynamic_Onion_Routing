package clienttui

import (
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	tea "github.com/charmbracelet/bubbletea"
)

func startClient(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		c.Run()
		return nil
	}
}

func waitEvent(ch <-chan client.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return doneMsg{}
		}
		return eventMsg{Event: ev}
	}
}
