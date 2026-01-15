package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showError {
			switch msg.String() {
			case "enter", "esc":
				m.showError = false
				m.err = nil
				return m, nil
			case "ctrl+c", "q":
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "ctrl+w":
			m.formFocused = !m.formFocused
			return m, nil

		case "alt+up":
			m.formFocused = true
			return m, nil

		case "alt+down":
			m.formFocused = false
			return m, nil
		}

		if m.formFocused {
			return m.updateFormKeys(msg)
		}
		return m.updateLogsKeys(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case logMsg:
		m.logSeq++
		m.logs = append(m.logs, logEntry{n: m.logSeq, s: msg.message})

		if len(m.logs) > m.maxLogs {
			m.logs = m.logs[len(m.logs)-m.maxLogs:]
		}

		maxOff := len(m.logs) - 1
		maxOff = max(maxOff, 0)
		if m.logsScrollOffset > maxOff {
			m.logsScrollOffset = maxOff
		}

	case errorMsg:
		m.err = msg.err
		m.showError = true
		m.processing = false
		return m, nil

	case doneMsg:
		m.processing = false
		m.inputs[2].value = ""

		m.inputs[0].focused = false
		m.inputs[1].focused = false
		m.inputs[2].focused = true
		m.focusIndex = 2
		m.cursorPos = 0

		return m, nil

	case cacheUpdateMsg:
		m.cachedMessage = msg.cachedMessage
		m.lastConfig = msg.lastConfig
		return m, nil
	}

	return m, nil
}

func (m tuiModel) updateFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "shift+tab", "up", "down":
		if !m.processing {
			s := msg.String()
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			for i := range m.inputs {
				m.inputs[i].focused = i == m.focusIndex
			}

			if m.focusIndex < len(m.inputs) {
				m.cursorPos = len(m.inputs[m.focusIndex].value)
			}
		}
		return m, nil

	case "enter":
		if !m.processing && m.focusIndex == len(m.inputs) {
			m.sink.config.Dest = m.inputs[0].value
			m.sink.config.OnionPath = m.inputs[1].value
			m.sink.config.Payload = m.inputs[2].value
			m.processing = true
			return m, m.startProcessing()
		}
		return m, nil

	case "left":
		if !m.processing && m.focusIndex < len(m.inputs) && m.cursorPos > 0 {
			m.cursorPos--
		}
		return m, nil

	case "right":
		if !m.processing && m.focusIndex < len(m.inputs) {
			if m.cursorPos < len(m.inputs[m.focusIndex].value) {
				m.cursorPos++
			}
		}
		return m, nil

	case "home", "ctrl+a":
		if !m.processing && m.focusIndex < len(m.inputs) {
			m.cursorPos = 0
		}
		return m, nil

	case "end", "ctrl+e":
		if !m.processing && m.focusIndex < len(m.inputs) {
			m.cursorPos = len(m.inputs[m.focusIndex].value)
		}
		return m, nil

	case "backspace":
		if !m.processing && m.focusIndex < len(m.inputs) && m.cursorPos > 0 {
			input := &m.inputs[m.focusIndex]
			input.value = input.value[:m.cursorPos-1] + input.value[m.cursorPos:]
			m.cursorPos--
		}
		return m, nil

	case "delete":
		if !m.processing && m.focusIndex < len(m.inputs) {
			input := &m.inputs[m.focusIndex]
			if m.cursorPos < len(input.value) {
				input.value = input.value[:m.cursorPos] + input.value[m.cursorPos+1:]
			}
		}
		return m, nil

	default:
		if !m.processing && m.focusIndex < len(m.inputs) && len(msg.String()) == 1 {
			input := &m.inputs[m.focusIndex]
			input.value = input.value[:m.cursorPos] + msg.String() + input.value[m.cursorPos:]
			m.cursorPos++
		}
		return m, nil
	}
}

func (m tuiModel) updateLogsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxOff := len(m.logs) - 1
	maxOff = max(maxOff, 0)

	switch msg.String() {
	case "up", "k":
		m.logsScrollOffset++
		if m.logsScrollOffset > maxOff {
			m.logsScrollOffset = maxOff
		}
		return m, nil

	case "down", "j":
		m.logsScrollOffset--
		if m.logsScrollOffset < 0 {
			m.logsScrollOffset = 0
		}
		return m, nil

	case "home":
		m.logsScrollOffset = maxOff
		return m, nil

	case "end":
		m.logsScrollOffset = 0
		return m, nil

	default:
		return m, nil
	}
}
