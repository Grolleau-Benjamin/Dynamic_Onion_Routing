package stdout

import (
	"fmt"
	"strings"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client/model"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/onion"
)

type Sink struct {
	client *client.Client
	config client.InputConfig
}

func New(c *client.Client, ic client.InputConfig) *Sink {
	return &Sink{
		client: c,
		config: ic,
	}
}

func (s *Sink) Start() error {
	done := make(chan struct{})

	msg, err := model.BuildFromInputConfig(s.config)
	if err != nil {
		return err
	}

	go func() {
		for ev := range s.client.Events() {
			switch ev.Type {
			case client.EvLog:
				fmt.Println("[LOG]", ev.Payload)
			case client.EvErr:
				fmt.Printf("[ERR] %v\n", ev.Payload)
			}
		}
		close(done)
	}()

	for gi := range msg.Path {
		group := &msg.Path[gi]
		if err = group.GenerateCryptoMaterial(); err != nil {
			return err
		}
		s.client.EmitLog(fmt.Sprintf("crypto material generated for %s", group))

		writeIdx := 0
		for ri := range group.Group.Relays {
			relay := &group.Group.Relays[ri]
			if err = s.client.RetrieveRelayIdentity(relay); err != nil {
				if strings.Contains(err.Error(), "connect: connection refused") {
					s.client.EmitLog(fmt.Sprintf("relay %s unreachable, skipping", relay.Ep.String()))
					continue
				}
				return err
			}
			if writeIdx != ri {
				group.Group.Relays[writeIdx] = group.Group.Relays[ri]
			}
			writeIdx++
		}
		group.Group.Relays = group.Group.Relays[:writeIdx]

		if len(group.Group.Relays) == 0 {
			return fmt.Errorf("no valid relays in group %d", gi)
		}
	}

	layer, err := onion.BuildOnion(msg.Dest, msg.Path, msg.Payload)
	if err != nil {
		return err
	}

	packet, err := layer.BytesPadded()
	if err != nil {
		return err
	}

	s.client.EmitLog(fmt.Sprintf("Sending a %d bytes packet", len(packet)))
	s.client.EmitLog(fmt.Sprintf("Packet prefix=%x", packet[:64]))

	sent := false
	for _, relay := range msg.Path[0].Group.Relays {
		entry := relay.Ep
		if err := s.client.SendOnionPacket(entry, packet); err != nil {
			s.client.EmitLog(fmt.Sprintf("failed to send onion packet to %s: %v", entry.String(), err))
			continue
		}
		s.client.EmitLog(fmt.Sprintf("onion packet sent to %s", entry.String()))
		sent = true
		break
	}
	if !sent {
		return fmt.Errorf("failed to send onion packet to any entry relay")
	}

	s.client.Close()
	<-done
	return nil
}
