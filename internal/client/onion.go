package client

import (
	"fmt"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/onion"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
)

func (c *Client) SendOnionPacket(ep identity.Endpoint, raw []byte) error {
	if len(raw) != onion.PacketSize {
		return fmt.Errorf("invalid onion packet size: got %d, want %d", len(raw), onion.PacketSize)
	}

	var pkt packet.OnionPacket
	copy(pkt.Data[:], raw)

	return c.SendPacket(ep, &pkt)
}
