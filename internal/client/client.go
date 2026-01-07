package client

import (
	"context"
)

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

func (c *Client) EmitLog(payload string) {
	c.events <- Event{
		Type:    EvLog,
		Payload: payload,
	}
}

func (c *Client) Events() <-chan Event {
	return c.events
}

func (c Client) Close() {
	close(c.events)
	c.cancel()
}
