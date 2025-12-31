package tui

import "strings"

func (m *Model) View() string {
	var sb strings.Builder

	sb.WriteString("Dynamic Onion Routing Client (Press 'q' to quit)\n\n")

	for _, l := range m.state.Logs {
		sb.WriteString(l)
		sb.WriteByte('\n')
	}

	return sb.String()
}
