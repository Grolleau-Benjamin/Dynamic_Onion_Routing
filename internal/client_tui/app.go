package clienttui

import (
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(c *client.Client, input client.InputConfig) error {
	uiEvents := make(chan client.Event, 32)

	go func() {
		for ev := range c.Events() {
			uiEvents <- ev
		}
		close(uiEvents)
	}()

	p := tea.NewProgram(
		NewModel(c, uiEvents, input),
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	return err
}
