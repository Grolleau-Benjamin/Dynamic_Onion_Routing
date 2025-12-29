package client

import (
	"time"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
)

type Client struct {
	events chan Event
}

func New() *Client {
	return &Client{
		events: make(chan Event, 64),
	}
}

func (c *Client) Events() <-chan Event {
	return c.events
}

func (c *Client) emit(ev Event) {
	select {
	case c.events <- ev:
	default: // drop cause noone listening
	}
}

func (c *Client) Run() {
	defer close(c.events)

	c.emit(Event{
		Type:    EventLog,
		Level:   logger.Info,
		Message: "Client started...",
		Time:    time.Now(),
	})

	time.Sleep(10 * time.Second)
}
