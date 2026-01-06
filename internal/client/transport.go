package client

import (
	"net"
	"time"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
)

func dialEndpoint(ep identity.Endpoint, timeout time.Duration) (net.Conn, error) {
	d := net.Dialer{
		Timeout: timeout,
	}
	return d.Dial(ep.Network(), ep.String())
}
