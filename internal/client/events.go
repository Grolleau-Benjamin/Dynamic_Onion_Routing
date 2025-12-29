package client

import (
	"time"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
)

type EventType int

const (
	EventLog EventType = iota
	EventError
)

type Event struct {
	Type    EventType
	Level   logger.Level
	Message string
	Time    time.Time
}
