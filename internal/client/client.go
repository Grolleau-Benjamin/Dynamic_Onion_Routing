package client

import (
	"context"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/transport"
)

type Client struct {
	events chan Event
	ctx    context.Context
	cancel context.CancelFunc

	tx *transport.Transport
}

func New() *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		events: make(chan Event, 10),
		ctx:    ctx,
		cancel: cancel,

		tx: transport.NewTransport(),
	}
}

func (c *Client) SendPacket(ep identity.Endpoint, p packet.Packet) error {
	return c.tx.Send(ep, p)
}

func (c *Client) RequestPacket(ep identity.Endpoint, req packet.Packet) (packet.Packet, error) {
	return c.tx.Request(ep, req)
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
