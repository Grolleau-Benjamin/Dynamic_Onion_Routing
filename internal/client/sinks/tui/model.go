package tui

import (
	"fmt"
	"strings"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client/model"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/onion"

	tea "github.com/charmbracelet/bubbletea"
)

type Sink struct {
	client *client.Client
	config client.InputConfig
}

func New(c *client.Client, ic client.InputConfig) *Sink {
	return &Sink{
		client: c,
		config: ic,
	}
}

type logMsg struct {
	message string
}

type errorMsg struct {
	err error
}

type doneMsg struct{}

type cacheUpdateMsg struct {
	cachedMessage *model.Message
	lastConfig    client.InputConfig
}

type inputField struct {
	label       string
	placeholder string
	value       string
	focused     bool
}

type logEntry struct {
	n int
	s string
}

type tuiModel struct {
	sink       *Sink
	inputs     []inputField
	focusIndex int

	logs   []logEntry
	logSeq int

	err       error
	showError bool

	width  int
	height int

	processing bool
	maxLogs    int

	cachedMessage *model.Message
	lastConfig    client.InputConfig

	formFocused      bool
	logsScrollOffset int

	cursorPos int
}

func initialModel(s *Sink) tuiModel {
	inputs := []inputField{
		{
			label:       "Destination",
			placeholder: "e.g., [cafe::1]:8080",
			value:       s.config.Dest,
			focused:     true,
		},
		{
			label:       "Onion Path",
			placeholder: "e.g., [::1]:62503,[::1]:62504|[::1]:62505",
			value:       s.config.OnionPath,
		},
		{
			label:       "Payload",
			placeholder: "e.g., test",
			value:       s.config.Payload,
		},
	}

	return tuiModel{
		sink:             s,
		inputs:           inputs,
		focusIndex:       0,
		logs:             []logEntry{},
		logSeq:           0,
		maxLogs:          50,
		processing:       false,
		formFocused:      true,
		logsScrollOffset: 0,
		cursorPos:        0,
	}
}

func (m tuiModel) Init() tea.Cmd {
	if m.sink.config.IsComplete() {
		return tea.Batch(tea.EnterAltScreen, m.startProcessing())
	}
	return tea.EnterAltScreen
}

func (m tuiModel) startProcessing() tea.Cmd {
	cachedMessage := m.cachedMessage
	lastConfig := m.lastConfig
	sinkConfig := m.sink.config

	return func() tea.Msg {
		configChanged := cachedMessage == nil ||
			sinkConfig.Dest != lastConfig.Dest ||
			sinkConfig.OnionPath != lastConfig.OnionPath

		payloadChanged := cachedMessage != nil && sinkConfig.Payload != lastConfig.Payload

		var msg model.Message
		var err error

		if configChanged {
			msg, err = model.BuildFromInputConfig(sinkConfig)
			if err != nil {
				return errorMsg{err: err}
			}

			for gi := range msg.Path {
				group := &msg.Path[gi]
				if err = group.GenerateCryptoMaterial(); err != nil {
					return errorMsg{err: err}
				}
				m.sink.client.EmitLog(fmt.Sprintf("crypto material generated for %s", group))

				writeIdx := 0
				for ri := range group.Group.Relays {
					relay := &group.Group.Relays[ri]
					if err = m.sink.client.RetrieveRelayIdentity(relay); err != nil {
						if strings.Contains(err.Error(), "connect: connection refused") {
							m.sink.client.EmitLog(fmt.Sprintf("relay %s unreachable, skipping", relay.Ep.String()))
							continue
						}
						return errorMsg{err: err}
					}
					if writeIdx != ri {
						group.Group.Relays[writeIdx] = group.Group.Relays[ri]
					}
					writeIdx++
				}
				group.Group.Relays = group.Group.Relays[:writeIdx]

				if len(group.Group.Relays) == 0 {
					return errorMsg{err: fmt.Errorf("no valid relays in group %d", gi)}
				}
			}

			cachedMessage = &msg
			lastConfig = sinkConfig
		} else if payloadChanged {
			msg = *cachedMessage
			msg.Payload = []byte(sinkConfig.Payload)
			m.sink.client.EmitLog("Reusing cached relay identities and crypto material (payload changed)")
			lastConfig.Payload = sinkConfig.Payload
		} else {
			msg = *cachedMessage
			m.sink.client.EmitLog("Reusing cached relay identities and crypto material")
		}

		layer, err := onion.BuildOnion(msg.Dest, msg.Path, msg.Payload)
		if err != nil {
			return errorMsg{err: err}
		}

		packet, err := layer.BytesPadded()
		if err != nil {
			return errorMsg{err: err}
		}

		m.sink.client.EmitLog(fmt.Sprintf("Sending a %d bytes packet", len(packet)))
		m.sink.client.EmitLog(fmt.Sprintf("Packet prefix=%x", packet[:24]))

		sent := false
		for _, relay := range msg.Path[0].Group.Relays {
			entry := relay.Ep
			if err := m.sink.client.SendOnionPacket(entry, packet); err != nil {
				m.sink.client.EmitLog(fmt.Sprintf("failed to send onion packet to %s: %v", entry.String(), err))
				continue
			}
			m.sink.client.EmitLog(fmt.Sprintf("onion packet sent to %s", entry.String()))
			sent = true
			break
		}

		if !sent {
			return errorMsg{err: fmt.Errorf("failed to send onion packet to any entry relay")}
		}

		return tea.Batch(
			func() tea.Msg {
				return cacheUpdateMsg{
					cachedMessage: cachedMessage,
					lastConfig:    lastConfig,
				}
			},
			func() tea.Msg { return doneMsg{} },
		)()
	}
}

func (s *Sink) Start() error {
	eventChan := make(chan tea.Msg, 100)

	go func() {
		for ev := range s.client.Events() {
			switch ev.Type {
			case client.EvLog:
				eventChan <- logMsg{message: fmt.Sprintf("%v", ev.Payload)}
			case client.EvErr:
				eventChan <- logMsg{message: fmt.Sprintf("[ERR] %v", ev.Payload)}
			}
		}
		close(eventChan)
	}()

	m := initialModel(s)

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
	)

	go func() {
		for msg := range eventChan {
			p.Send(msg)
		}
	}()

	_, err := p.Run()
	s.client.Close()
	return err
}
