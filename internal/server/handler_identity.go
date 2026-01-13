package server

import (
	"net"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
)

func handleGetIdentity(p packet.Packet, conn net.Conn, s *Server) {
	logger.Debugf("[%s] GetIdentityRequest received", conn.RemoteAddr())

	resp := &packet.GetIdentityResponse{
		Ruuid:     s.Pi.UUID,
		PublicKey: s.Pi.PubKey,
	}

	if err := packet.WritePacket(conn, resp); err != nil {
		logger.Warnf("[%s] failed to send identity response: %v", conn.RemoteAddr(), err)
	}
}
