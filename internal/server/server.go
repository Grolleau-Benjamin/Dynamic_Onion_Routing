package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
)

type Server struct {
	ln net.Listener

	ep identity.Endpoint
	pi *identity.PrivateIdentity

	wg       sync.WaitGroup
	stop     chan struct{}
	stopOnce sync.Once
}

func New(addr, idDir string, port uint16) (*Server, error) {
	ep, err := identity.NewEndpoint(addr, port)
	if err != nil {
		return nil, err
	}

	pi, err := identity.LoadPrivateIdentity(idDir)
	if err != nil {
		return nil, err
	}

	ln, err := net.Listen(ep.Network(), ep.String())
	if err != nil {
		return nil, err
	}

	return &Server{
		ln: ln,

		ep: ep,
		pi: pi,

		stop: make(chan struct{}),
	}, nil
}

func (s *Server) Serve(ctx context.Context) error {
	fmt.Println("Server just started")

	errCh := make(chan error, 1)
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		for {
			conn, err := s.ln.Accept()
			if err != nil {
				// Check for normal shutdown (Listener closed)
				if errors.Is(err, net.ErrClosed) {
					errCh <- nil
					return
				}

				// Check if we were asked to stop via channel
				select {
				case <-s.stop:
					errCh <- nil
					return
				default:
					time.Sleep(100 * time.Millisecond)
					continue
				}
			}

			s.wg.Add(1)
			go s.handleConn(conn)
		}
	}()

	select {
	// Context cancelled via Signal (Ctrl+C)
	case <-ctx.Done():
		_ = s.close()
		return ctx.Err()

	// Internal error or Listener close
	case err := <-errCh:
		_ = s.close()
		return err
	}
}

func (s *Server) close() error {
	var err error
	s.stopOnce.Do(func() {
		close(s.stop)
		if s.ln != nil {
			err = s.ln.Close()
		}
		s.wg.Wait()
	})
	return err
}
