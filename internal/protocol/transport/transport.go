package transport

import (
	"net"
	"time"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
)

type Transport struct {
	dialTimeout  time.Duration
	writeTimeout time.Duration
	readTimeout  time.Duration
}

func dialEndpoint(ep identity.Endpoint, timeout time.Duration) (net.Conn, error) {
	d := net.Dialer{Timeout: timeout}
	return d.Dial(ep.Network(), ep.String())
}

func NewTransport() *Transport {
	return &Transport{
		dialTimeout:  3 * time.Second,
		writeTimeout: 2 * time.Second,
		readTimeout:  5 * time.Second,
	}
}

func (t *Transport) dial(ep identity.Endpoint) (net.Conn, error) {
	return dialEndpoint(ep, t.dialTimeout)
}

func (t *Transport) Send(ep identity.Endpoint, p packet.Packet) error {
	conn, err := t.dial(ep)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	if err := conn.SetWriteDeadline(time.Now().Add(t.writeTimeout)); err != nil {
		return err
	}

	return packet.WritePacket(conn, p)
}

func (t *Transport) Request(ep identity.Endpoint, req packet.Packet) (packet.Packet, error) {
	conn, err := t.dial(ep)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()

	if err = conn.SetWriteDeadline(time.Now().Add(t.writeTimeout)); err != nil {
		return nil, err
	}
	if err = packet.WritePacket(conn, req); err != nil {
		return nil, err
	}

	if err = conn.SetReadDeadline(time.Now().Add(t.readTimeout)); err != nil {
		return nil, err
	}
	resp, err := packet.ReadPacket(conn)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
