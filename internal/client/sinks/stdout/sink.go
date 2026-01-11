package stdout

import (
	"fmt"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client/model"
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

	fmt.Println(msg)

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

		if err := group.GenerateCryptoMateriel(); err != nil {
			return err
		}
		s.client.EmitLog(fmt.Sprintf("crypto material generated for %s", group))

		for ri := range group.Group.Relays {
			relay := &group.Group.Relays[ri]

			if err := s.client.RetrieveRelayIdentity(relay); err != nil {
				return err
			}
		}

		// ol := model.BuildFromInputConfig(ic client.InputConfig)
	}

	s.client.Close()
	<-done
	return nil
}
