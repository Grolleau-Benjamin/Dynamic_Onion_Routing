package client

import (
	"fmt"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
)

func (c *Client) RetrieveRelayIdentity(r *identity.Relay) error {
	resp, err := c.tx.Request(r.Ep, &packet.GetIdentityRequest{})
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
