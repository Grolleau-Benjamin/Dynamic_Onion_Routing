package server

import (
	"errors"
	"io"
	"net"
	"sync"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/logger"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
)

type HandlerFunc func(
	p packet.Packet,
	conn net.Conn,
	s *Server,
)

var handlerRegistry = map[uint8]HandlerFunc{
	packet.TypeGetIdentityRequest: handleGetIdentity,
	packet.TypeOnionPacket:        handleOnionPacket,
}

func (s *Server) handleConn(conn net.Conn) {
	var wg sync.WaitGroup

	defer func() {
		wg.Wait()

		if err := conn.Close(); err != nil {
			logger.Warnf("error closing connection: %v", err)
		}
	}()

	remote := conn.RemoteAddr().String()

	for {
		pkt, err := packet.ReadPacket(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return // remote closed the connection
			}
			logger.Warnf("[%s] read packet failed: %v", remote, err)
			return
		}

		h, ok := handlerRegistry[pkt.Type()]
		if !ok {
			logger.Warnf("[%s] no handler for packet type 0x%02x", remote, pkt.Type())
			return
		}

		wg.Add(1)
		go func(p packet.Packet) {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("[%s] PANIC in handler: %v", remote, r)
				}
			}()

			h(p, conn, s)
		}(pkt)
	}
}
