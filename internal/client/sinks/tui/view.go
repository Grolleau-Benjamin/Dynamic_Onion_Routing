package tui

import (
	"fmt"
	"strings"
)

func (m tuiModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var b strings.Builder

	title := titleStyle.Width(m.width).Render("🧅 Dynamic Onion Routing Client")
	b.WriteString(title)
	b.WriteString("\n\n")

	titleHeight := 3
	helpHeight := 2
	spacingHeight := 1
	availableHeight := m.height - titleHeight - helpHeight - spacingHeight
	availableHeight = max(availableHeight, 15)

	formContentHeight := int(float64(availableHeight) * 0.60)
	logsContentHeight := int(float64(availableHeight) * 0.35)

	total := formContentHeight + logsContentHeight
	if total > availableHeight {
		excess := total - availableHeight
		logsContentHeight -= excess
	}

	if formContentHeight < 8 {
		formContentHeight = 8
	}
	if logsContentHeight < 4 {
		logsContentHeight = 4
	}

	maxLogsHeight := int(float64(m.height) * 0.30)
	if logsContentHeight > maxLogsHeight {
		logsContentHeight = maxLogsHeight
	}

	formStyle := inactiveFormBoxStyle
	logsStyle := inactiveLogsBoxStyle
	if m.formFocused {
		formStyle = activeFormBoxStyle
	} else {
		logsStyle = activeLogsBoxStyle
	}

	formContent := m.renderForm()
	formBox := formStyle.Width(m.width - 4).Height(formContentHeight).Render(formContent)
	b.WriteString(formBox)
	b.WriteString("\n")

	logsContent := m.renderLogs(logsContentHeight)
	logsBox := logsStyle.Width(m.width - 4).Height(logsContentHeight).Render(logsContent)
	b.WriteString(logsBox)
	b.WriteString("\n")

	var helpText string
	if m.formFocused {
		helpText = "↑/↓ or Tab: Navigate | ←/→: Move cursor | Alt+↓: Logs | Enter: Submit | Ctrl+C: Quit"
		if m.processing {
			helpText = "Processing... | Alt+↓: Logs | Ctrl+C: Quit"
		}
	} else {
		helpText = "↑/↓: Scroll | Home/End: Oldest/Newest | Alt+↑: Form | Ctrl+C: Quit"
	}
	b.WriteString(helpStyle.Width(m.width).Render(helpText))

	if m.showError {
		return m.renderWithErrorPopup(b.String())
	}

	return b.String()
}

func (m tuiModel) renderForm() string {
	var b strings.Builder

	header := "Configuration"
	b.WriteString(blurredStyle.Render(header))
	b.WriteString("\n\n")

	for _, input := range m.inputs {
		if input.focused && !m.processing && m.formFocused {
			b.WriteString(focusedStyle.Render("> " + input.label + ":"))
			b.WriteString("\n")

			beforeCursor := input.value[:m.cursorPos]
			afterCursor := input.value[m.cursorPos:]
			b.WriteString(focusedStyle.Render("  " + beforeCursor))
			b.WriteString(cursorStyle.Render("▌"))
			b.WriteString(focusedStyle.Render(afterCursor))
			b.WriteString("\n")

			if input.value == "" {
				b.WriteString(blurredStyle.Render("  " + input.placeholder))
				b.WriteString("\n")
			}
		} else {
			if m.formFocused {
				b.WriteString(blurredStyle.Render("  " + input.label + ":"))
			} else {
				b.WriteString(noStyle.Render("  " + input.label + ":"))
			}
			b.WriteString("\n")

			if input.value != "" {
				b.WriteString(noStyle.Render("  " + input.value))
			} else {
				b.WriteString(blurredStyle.Render("  " + input.placeholder))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	button := blurredButton
	if m.focusIndex == len(m.inputs) && !m.processing && m.formFocused {
		button = focusedButton
	} else if m.processing {
		button = blurredStyle.Render("[ Processing... ]")
	}
	b.WriteString(button)

	return b.String()
}

func (m tuiModel) renderLogs(boxHeight int) string {
	var b strings.Builder

	header := "Logs"
	if !m.formFocused {
		header += " [ACTIVE]"
	}

	if m.processing {
		b.WriteString(infoStyle.Render("⟳ " + header))
	} else if len(m.logs) > 0 {
		b.WriteString(successStyle.Render("✓ " + header))
	} else {
		if !m.formFocused {
			b.WriteString(focusedStyle.Render(header))
		} else {
			b.WriteString(blurredStyle.Render(header))
		}
	}
	b.WriteString("\n\n")

	if len(m.logs) == 0 {
		b.WriteString(blurredStyle.Render("No logs yet. Fill the form and press Submit to start."))
		return b.String()
	}

	maxLogLines := boxHeight - 6
	maxLogLines = max(maxLogLines, 2)

	contentWidth := m.width - 16
	contentWidth = max(contentWidth, 40)

	scrollOffset := m.logsScrollOffset
	maxOff := len(m.logs) - 1
	maxOff = max(maxOff, 0)
	scrollOffset = max(scrollOffset, 0)
	scrollOffset = min(scrollOffset, maxOff)

	startIdx := len(m.logs) - 1 - scrollOffset
	startIdx = max(startIdx, 0)
	if startIdx >= len(m.logs) {
		startIdx = len(m.logs) - 1
	}

	var visible []logEntry
	linesUsed := 0

	for i := startIdx; i >= 0; i-- {
		e := m.logs[i]
		display := fmt.Sprintf("%5d │ %s", e.n, e.s)
		w := len([]rune(display))
		estimatedLines := (w / contentWidth) + 1

		if linesUsed+estimatedLines > maxLogLines {
			break
		}

		visible = append([]logEntry{e}, visible...)
		linesUsed += estimatedLines
	}

	if len(visible) == 0 {
		visible = []logEntry{m.logs[len(m.logs)-1]}
	}

	for _, e := range visible {
		b.WriteString(m.formatLog(e.n, e.s))
		b.WriteString("\n")
	}

	firstIdx := startIdx - (len(visible) - 1)
	firstIdx = max(firstIdx, 0)
	hiddenAbove := firstIdx
	hiddenBelow := (len(m.logs) - 1) - startIdx

	if hiddenAbove > 0 || hiddenBelow > 0 {
		b.WriteString("\n")
		b.WriteString(blurredStyle.Render(fmt.Sprintf("↑ %d above | %d below ↓", hiddenAbove, hiddenBelow)))
	}

	return b.String()
}

func (m tuiModel) renderWithErrorPopup(content string) string {
	popupWidth := 60
	if m.width < 70 {
		popupWidth = m.width - 10
	}

	errorContent := errorStyle.Render("Error") + "\n\n" +
		fmt.Sprintf("%v", m.err) + "\n\n" +
		helpStyle.Render("Press Enter or Esc to close")

	popup := errorPopupStyle.Width(popupWidth).Render(errorContent)

	lines := strings.Split(content, "\n")

	popupHeight := strings.Count(popup, "\n") + 3
	startLine := (len(lines) - popupHeight) / 2
	startLine = max(startLine, 0)

	var result strings.Builder
	for i, line := range lines {
		if i == startLine {
			padding := (m.width - popupWidth) / 2
			padding = max(padding, 0)
			for pLine := range strings.SplitSeq(popup, "\n") {
				result.WriteString(strings.Repeat(" ", padding))
				result.WriteString(pLine)
				result.WriteString("\n")
			}
		} else if i > startLine && i < startLine+popupHeight {
			continue
		} else {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String()
}

func (m tuiModel) formatLog(n int, log string) string {
	prefix := lineNoStyle.Render(fmt.Sprintf("%5d │ ", n))

	if strings.Contains(log, "generated") {
		return prefix + successStyle.Render("✓ "+log)
	} else if strings.Contains(log, "identity received") {
		return prefix + infoStyle.Render("→ "+log)
	} else if strings.Contains(log, "Sending") {
		return prefix + successStyle.Render("📤 "+log)
	} else if strings.Contains(log, "sent to") {
		return prefix + successStyle.Render("✓ "+log)
	} else if strings.Contains(log, "failed") {
		return prefix + errorStyle.Render("✗ "+log)
	}
	return prefix + logStyle.Render(log)
}
