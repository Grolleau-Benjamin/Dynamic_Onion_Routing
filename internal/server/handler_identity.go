package server

import (
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
)

func handleGetIdentity(
	p packet.Packet,
	s *Server,
) (packet.Packet, error) {
	logger.Debugf("Requesting for GetIdentity")

	return &packet.GetIdentityResponse{
		Ruuid:     s.Pi.UUID,
		PublicKey: s.Pi.PubKey,
	}, nil
}
