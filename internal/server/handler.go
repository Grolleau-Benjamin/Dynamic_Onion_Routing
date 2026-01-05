package server

import (
	"fmt"
	"net"
	"time"
)

func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		fmt.Printf("Failed to set deadline: %v\n", err)
		return
	}

	remoteAddr := conn.RemoteAddr().String()
	fmt.Printf("[%s] Connection accepted\n", remoteAddr)
}
