package server

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
)

type HandlerFunc func(
	p packet.Packet,
	s *Server,
) (packet.Packet, error)

var handlerRegistry = map[uint8]HandlerFunc{
	packet.TypeGetIdentityRequest: handleGetIdentity,
}

func (s *Server) handleConn(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Warnf("error closing connection: %v", err)
		}
	}()

	remote := conn.RemoteAddr().String()
	logger.Debugf("[%s] connection accepted", remote)

	for {
		if err := conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
			logger.Warnf("[%s] failed to set deadline: %v", remote, err)
			return
		}

		pkt, err := packet.ReadPacket(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Debugf("[%s] connection closed by peer", remote)
				return
			}
			if errors.Is(err, io.ErrUnexpectedEOF) {
				logger.Warnf("[%s] truncated packet", remote)
				return
			}
			logger.Warnf("[%s] read packet failed: %v", remote, err)
			return
		}

		h, ok := handlerRegistry[pkt.Type()]
		if !ok {
			logger.Warnf("[%s] no handler for packet type 0x%02x", remote, pkt.Type())
			return
		}

		resp, err := h(pkt, s)
		if err != nil {
			logger.Warnf("[%s] handler error: %v", remote, err)
			return
		}

		if resp != nil {
			if err := packet.WritePacket(conn, resp); err != nil {
				logger.Warnf("[%s] write response failed: %v", remote, err)
				return
			}
		}
	}
}
