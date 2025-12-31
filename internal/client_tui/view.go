package clienttui

import (
	"fmt"
	"strings"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.state.Width == 0 || m.state.Height == 0 {
		return ""
	}

	var tabHome, tabSelect string
	if m.state.Mode == ModeHome {
		tabHome = TabActiveStyle.Render("1. Home")
		tabSelect = TabInactiveStyle.Render("2. Select")
	} else {
		tabHome = TabInactiveStyle.Render("1. Home")
		tabSelect = TabActiveStyle.Render("2. Select")
	}

	header := lipgloss.JoinHorizontal(lipgloss.Top, tabHome, tabSelect)

	footer := lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderKey("esc", "quit"),
		renderKey("1", "home"),
		renderKey("2", "select"),
	)

	availableHeight := m.state.Height - lipgloss.Height(header) - lipgloss.Height(footer)
	hFrame, vFrame := ContainerStyle.GetFrameSize()

	innerWidth := m.state.Width - hFrame
	innerHeight := availableHeight - vFrame

	var content string
	switch m.state.Mode {
	case ModeHome:
		content = homeView(m, innerWidth, innerHeight)
	case ModeSelection:
		content = selectView(m)
	}

	mainBox := lipgloss.NewStyle().
		Width(innerWidth).
		Height(innerHeight).
		Render(content)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		mainBox,
		footer,
	)
}

func renderKey(key, desc string) string {
	return fmt.Sprintf("%s %s   ", KeyStyle.Render("<"+key+">"), DescStyle.Render(desc))
}

func homeView(m Model, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	topH := height / 2
	bottomH := height - topH

	hFrame, vFrame := SubContainerStyle.GetFrameSize()
	subWidth := width - hFrame

	topBox := SubContainerStyle.
		Width(subWidth).
		Height(topH - vFrame).
		Render(networkView(m.dest, m.path))

	bottomBox := SubContainerStyle.
		Width(subWidth).
		Height(bottomH - vFrame).
		Render(logView(m.events, bottomH-vFrame))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topBox,
		bottomBox,
	)
}

func networkView(dest identity.Endpoint, path []client.GroupCryptoContext) string {
	var b strings.Builder

	for i, hop := range path {
		fmt.Fprintf(&b, "[HOP %d]\n", i+1)

		for j, r := range hop.Group.Relays {
			prefix := " ├─"
			if j == len(hop.Group.Relays)-1 {
				prefix = " └─"
			}

			uuid := fmt.Sprintf("%x", r.UUID[:4])

			fmt.Fprintf(
				&b,
				"%s %s (%s…)\n",
				prefix,
				r.Ep.String(),
				uuid,
			)
		}

		b.WriteString("\n")
	}

	b.WriteString("[DEST]\n")
	fmt.Fprintf(
		&b,
		" └─ %s\n",
		dest.String(),
	)

	return b.String()
}

func logView(events []client.Event, height int) string {
	if height <= 0 {
		return ""
	}

	start := 0
	if len(events) > height {
		start = len(events) - height
	}

	var b strings.Builder
	for _, ev := range events[start:] {
		fmt.Fprintln(&b, ev.String())
	}

	return b.String()
}

func selectView(m Model) string {
	return "Make a selection here."
}
