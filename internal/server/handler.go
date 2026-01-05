package server

import (
	"net"
	"time"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
)

func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()

	if err := conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		logger.Warnf("Failed to set deadline for %s: %v", remoteAddr, err)
		return
	}

	logger.Debugf("[%s] Connection accepted", remoteAddr)
}
