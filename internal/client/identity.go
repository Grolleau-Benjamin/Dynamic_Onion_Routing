package client

import (
	"fmt"
	"time"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
)

func (c *Client) RetrieveRelayIdentity(r *identity.Relay) error {
	conn, err := dialEndpoint(r.Ep, 5*time.Second)
	if err != nil {
		return err
	}
	defer func() {
		if err = conn.Close(); err != nil {
			c.events <- Event{Type: EvErr, Payload: err.Error()}
		}
	}()

	if err = conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return err
	}

	if err = packet.WritePacket(conn, &packet.GetIdentityRequest{}); err != nil {
		return err
	}

	resp, err := packet.ReadPacket(conn)
	if err != nil {
		return err
	}

	id, ok := resp.(*packet.GetIdentityResponse)
	if !ok {
		return fmt.Errorf("unexpected packet type %T", resp)
	}

	r.HydrateIdentity(id.Ruuid, id.PublicKey)

	c.EmitLog(fmt.Sprintf(
		"identity received from %s => uuid=%X pub=%X",
		r.Ep.String(),
		id.Ruuid[:4],
		id.PublicKey[:4],
	))

	return nil
}
