package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
)

type Client struct {
	events chan Event
	ctx    context.Context
	cancel context.CancelFunc
}

func New() *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		events: make(chan Event, 64),
		ctx:    ctx,
		cancel: cancel,
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

func (c *Client) Close() {
	c.cancel()
	close(c.events)
}

func (c *Client) Send(ctx SendContext) error {
	c.emit(Event{
		Type:    EventLog,
		Level:   logger.Debug,
		Message: fmt.Sprintf("Send(ctx) -> ctx = %v", ctx),
		Time:    time.Now(),
	})

	select {
	case <-c.ctx.Done():
		return fmt.Errorf("client closed")
	default:
	}
	c.emit(Event{
		Type:    EventLog,
		Level:   logger.Info,
		Message: "Client sending started...",
		Time:    time.Now(),
	})

	c.emit(Event{
		Type:    EventLog,
		Level:   logger.Debug,
		Message: "Building onion packet",
		Time:    time.Now(),
	})

	return nil
}

// func (c *Client) Run() {
// 	defer close(c.events)

// 	c.emit(Event{
// 		Type:    EventLog,
// 		Level:   logger.Info,
// 		Message: "Client started...",
// 		Time:    time.Now(),
// 	})

// 	time.Sleep(10 * time.Second)
// }
