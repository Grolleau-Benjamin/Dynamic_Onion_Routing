package clienttui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case eventMsg:
		m.events = append(m.events, msg.Event)
		return m, waitEvent(m.eventsCh)

	case doneMsg:
		m.done = true
		return m, nil
	}

	return m, nil
}
