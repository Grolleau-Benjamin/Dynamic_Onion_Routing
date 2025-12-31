package client

import "context"

type Client struct {
	events chan Event
	ctx    context.Context
	cancel context.CancelFunc
}

func New() *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		events: make(chan Event, 10),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Client) Events() <-chan Event {
	return c.events
}

// TODO: delete this later when no more usefull
func (c *Client) Simulate() {
	c.events <- Event{Type: EvLog, Payload: "[Simulating] Connecting..."}
	c.events <- Event{Type: EvLog, Payload: "[Simulating] Simulating data..."}
	c.events <- Event{Type: EvLog, Payload: "[Simulating] Done."}
}

func (c Client) Close() {
	close(c.events)
	c.cancel()
}
