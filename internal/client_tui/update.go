package clienttui

import (
	"time"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.state.Width = msg.Width
		m.state.Height = msg.Height
		return m, waitEvent(m.eventsCh)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.client.Close()
			return m, tea.Quit

		case tea.KeyEnter:
			m.events = append(m.events, client.Event{
				Type:    client.EventLog,
				Level:   logger.Info,
				Message: "ENTRER key has been pressed",
				Time:    time.Now(),
			})
			return m, waitEvent(m.eventsCh)

		}
		switch msg.String() {
		case "1":
			m.state.Mode = ModeHome
			return m, waitEvent(m.eventsCh)
		case "2":
			m.state.Mode = ModeSelection
			return m, waitEvent(m.eventsCh)
		}

	case eventMsg:
		m.events = append(m.events, msg.Event)
		const maxEvents = 200
		if len(m.events) > maxEvents {
			m.events = m.events[len(m.events)-maxEvents:]
		}
		return m, waitEvent(m.eventsCh)

	case doneMsg:
		m.done = true
		return m, nil
	}

	return m, waitEvent(m.eventsCh)
}
