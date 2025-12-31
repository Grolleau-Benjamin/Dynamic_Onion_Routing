package client

import (
	"fmt"
	"time"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
)

type EventType int

const (
	EventLog EventType = iota
	EventError
)

func (et EventType) String() string {
	switch et {
	case EventLog:
		return "ELog"
	case EventError:
		return "EErr"
	default:
		return "UnknownEvent"
	}
}

type Event struct {
	Type    EventType
	Level   logger.Level
	Message string
	Time    time.Time
}

func (e Event) String() string {
	return fmt.Sprintf(
		"[%s - %s - %s] %s",
		e.Time.Format("15:04:05"),
		e.Type,
		e.Level,
		e.Message,
	)
}
