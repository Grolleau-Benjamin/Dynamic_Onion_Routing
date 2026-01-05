package identity

import (
	"fmt"
	"net"
	"strconv"
)

type Endpoint struct {
	IP   net.IP
	Port uint16
}

func NewEndpoint(ipStr string, port uint16) (Endpoint, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return Endpoint{}, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	if port == 0 {
		return Endpoint{}, fmt.Errorf("invalid port: %d", port)
	}

	return Endpoint{
		IP:   ip,
		Port: port,
	}, nil
}

func ParseEpFromString(ep string) (Endpoint, error) {
	host, portStr, err := net.SplitHostPort(ep)
	if err != nil {
		return Endpoint{}, fmt.Errorf("invalid endpoint format %q: %w", ep, err)
	}

	port64, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return Endpoint{}, fmt.Errorf("invalid port %q: %w", portStr, err)
	}

	if port64 == 0 {
		return Endpoint{}, fmt.Errorf("invalid port: %d", port64)
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return Endpoint{}, fmt.Errorf("invalid IP address: %s", host)
	}

	return Endpoint{
		IP:   ip,
		Port: uint16(port64),
	}, nil
}

func (e Endpoint) String() string {
	if e.IP.To4() != nil {
		return fmt.Sprintf("%s:%d", e.IP.String(), e.Port)
	}
	return fmt.Sprintf("[%s]:%d", e.IP.String(), e.Port)
}

func (e Endpoint) IsIPv4() bool {
	return e.IP.To4() != nil
}

func (e Endpoint) IsIPv6() bool {
	return e.IP.To16() != nil && e.IP.To4() == nil
}

func (e Endpoint) Network() string {
	if e.IsIPv4() {
		return "tcp4"
	}
	return "tcp6"
}

