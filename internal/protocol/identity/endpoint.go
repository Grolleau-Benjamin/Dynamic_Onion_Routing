package identity

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type Endpoint struct {
	IP   net.IP
	Port uint16
}

const (
	EndpointIPv4 = 0x04
	EndpointIPv6 = 0x06
)

func (e Endpoint) Bytes() ([]byte, error) {
	var ipBytes []byte
	var ipType byte

	if ip4 := e.IP.To4(); ip4 != nil {
		ipType = EndpointIPv4
		ipBytes = ip4
	} else if ip6 := e.IP.To16(); ip6 != nil {
		ipType = EndpointIPv6
		ipBytes = ip6
	} else {
		return nil, fmt.Errorf("invalid IP address: %v", e.IP)
	}

	out := make([]byte, 1+2+len(ipBytes))
	out[0] = ipType
	binary.BigEndian.PutUint16(out[1:3], e.Port)
	copy(out[3:], ipBytes)

	return out, nil
}

func (e Endpoint) BytesLen() int {
	if ip4 := e.IP.To4(); ip4 != nil {
		return 7
	}
	if len(e.IP) > 0 {
		return 19
	}
	return 0
}

func (e *Endpoint) Parse(data []byte) (int, error) {
	if len(data) < 3 {
		return 0, fmt.Errorf("data too short")
	}

	ipType := data[0]
	e.Port = binary.BigEndian.Uint16(data[1:3])

	var n int
	switch ipType {
	case EndpointIPv4:
		n = 7
		if len(data) < n {
			return 0, fmt.Errorf("buffer too short for IPv4")
		}
		e.IP = make(net.IP, net.IPv4len)
		copy(e.IP, data[3:7])
	case EndpointIPv6:
		n = 19
		if len(data) < n {
			return 0, fmt.Errorf("buffer too short for IPv6")
		}
		e.IP = make(net.IP, net.IPv6len)
		copy(e.IP, data[3:19])
	default:
		return 0, fmt.Errorf("unknown IP type: 0x%02x", ipType)
	}

	return n, nil
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
