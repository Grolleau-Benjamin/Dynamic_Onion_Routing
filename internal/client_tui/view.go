package clienttui

import (
	"fmt"
	"strings"
)

func (m Model) View() string {
	var b strings.Builder

	b.WriteString("Dynamic Onion Routing - Events\n\n")

	for _, ev := range m.events {
		b.WriteString(fmt.Sprintf("[%v] %s\n", ev.Level, ev.Message))
	}

	if m.done {
		b.WriteString("\n(client finished)\n")
	}

	b.WriteString("\n[esc / ctrl+c] quit\n")
	return b.String()
}
