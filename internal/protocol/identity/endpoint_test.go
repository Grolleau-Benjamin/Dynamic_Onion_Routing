package identity_test

import (
	"bytes"
	"net"
	"strings"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
)

func TestEndpoint_Bytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ep          identity.Endpoint
		expected    []byte
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid IPv4",
			ep:          identity.Endpoint{IP: net.ParseIP("8.8.8.8"), Port: 62503},
			expected:    []byte{0x04, 0xf4, 0x27, 0x08, 0x08, 0x08, 0x08},
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "valid IPv6",
			ep:          identity.Endpoint{IP: net.ParseIP("::1"), Port: 62503},
			expected:    []byte{0x06, 0xf4, 0x27, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "nil IP",
			ep:          identity.Endpoint{IP: nil, Port: 62503},
			expected:    nil,
			wantErr:     true,
			errContains: "invalid IP address",
		},
		{
			name:        "empty IP",
			ep:          identity.Endpoint{IP: net.IP{}, Port: 8080},
			expected:    nil,
			wantErr:     true,
			errContains: "invalid IP address",
		},
		{
			name:        "zero-length IP",
			ep:          identity.Endpoint{IP: make(net.IP, 0), Port: 9000},
			expected:    nil,
			wantErr:     true,
			errContains: "invalid IP address",
		},
		{
			name:        "IPv4 with port 0",
			ep:          identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 0},
			expected:    []byte{0x04, 0x00, 0x00, 0xc0, 0xa8, 0x01, 0x01},
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "IPv6 with port 65535",
			ep:          identity.Endpoint{IP: net.ParseIP("fe80::1"), Port: 65535},
			expected:    []byte{0x06, 0xff, 0xff, 0xfe, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "IPv4 localhost",
			ep:          identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 80},
			expected:    []byte{0x04, 0x00, 0x50, 0x7f, 0x00, 0x00, 0x01},
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "IPv4-mapped IPv6",
			ep:          identity.Endpoint{IP: net.ParseIP("::ffff:192.0.2.1"), Port: 8080},
			expected:    []byte{0x04, 0x1f, 0x90, 0xc0, 0x00, 0x02, 0x01},
			wantErr:     false,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.ep.Bytes()

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("error mismatch:\n\tgot:  %q\n\twant: %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !bytes.Equal(got, tt.expected) {
				t.Fatalf("mismatch:\n\tgot:  %v\n\twant: %v", got, tt.expected)
			}
		})
	}
}

func TestEndpoint_BytesLen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ep       identity.Endpoint
		expected int
	}{
		{
			name:     "IPv4",
			ep:       identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 62503},
			expected: 7,
		},
		{
			name:     "IPv4 loopback",
			ep:       identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 62503},
			expected: 7,
		},
		{
			name:     "IPv6",
			ep:       identity.Endpoint{IP: net.ParseIP("2003:2506::1"), Port: 62503},
			expected: 19,
		},
		{
			name:     "IPv6 loopback",
			ep:       identity.Endpoint{IP: net.ParseIP("::1"), Port: 62503},
			expected: 19,
		},
		{
			name:     "IPv4-mapped IPv6",
			ep:       identity.Endpoint{IP: net.ParseIP("::ffff:192.168.1.1"), Port: 62503},
			expected: 7,
		},
		{
			name:     "nil IP",
			ep:       identity.Endpoint{IP: nil, Port: 62503},
			expected: 0,
		},
		{
			name:     "empty IP",
			ep:       identity.Endpoint{IP: net.IP{}, Port: 62503},
			expected: 0,
		},
		{
			name:     "zero-length IP",
			ep:       identity.Endpoint{IP: make(net.IP, 0), Port: 62503},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.ep.BytesLen()

			if got != tt.expected {
				t.Fatalf("mismatch:\n\tgot:  %d\n\twant: %d", got, tt.expected)
			}
		})
	}
}

func TestEndpoint_Parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		b           []byte
		expected    int
		expectedEp  identity.Endpoint
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid IPv4",
			b:        []byte{0x04, 0xf4, 0x27, 0xc0, 0xa8, 0x01, 0x01},
			expected: 7,
			expectedEp: identity.Endpoint{
				IP:   net.ParseIP("192.168.1.1"),
				Port: 62503,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name:     "valid IPv6",
			b:        []byte{0x06, 0xf4, 0x27, 0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			expected: 19,
			expectedEp: identity.Endpoint{
				IP:   net.ParseIP("2001:db8::1"),
				Port: 62503,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name:     "IPv4 localhost",
			b:        []byte{0x04, 0x00, 0x50, 0x7f, 0x00, 0x00, 0x01},
			expected: 7,
			expectedEp: identity.Endpoint{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 80,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name:     "IPv6 localhost",
			b:        []byte{0x06, 0x1f, 0x90, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			expected: 19,
			expectedEp: identity.Endpoint{
				IP:   net.ParseIP("::1"),
				Port: 8080,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "buffer too short",
			b:           []byte{0x04, 0xf4},
			expected:    0,
			expectedEp:  identity.Endpoint{},
			wantErr:     true,
			errContains: "data too short",
		},
		{
			name:        "empty buffer",
			b:           []byte{},
			expected:    0,
			expectedEp:  identity.Endpoint{},
			wantErr:     true,
			errContains: "data too short",
		},
		{
			name:        "invalid IP type",
			b:           []byte{0x99, 0xf4, 0x27, 0xc0, 0xa8, 0x01, 0x01},
			expected:    0,
			expectedEp:  identity.Endpoint{},
			wantErr:     true,
			errContains: "unknown IP type",
		},
		{
			name:     "IPv4 with port 0",
			b:        []byte{0x04, 0x00, 0x00, 0xc0, 0xa8, 0x01, 0x01},
			expected: 7,
			expectedEp: identity.Endpoint{
				IP:   net.ParseIP("192.168.1.1"),
				Port: 0,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name:     "IPv6 with port 65535",
			b:        []byte{0x06, 0xff, 0xff, 0xfe, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			expected: 19,
			expectedEp: identity.Endpoint{
				IP:   net.ParseIP("fe80::1"),
				Port: 65535,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "IPv4 buffer too short",
			b:           []byte{0x04, 0xf4, 0x27, 0xc0, 0xa8, 0x01},
			expected:    0,
			expectedEp:  identity.Endpoint{},
			wantErr:     true,
			errContains: "buffer too short",
		},
		{
			name:        "IPv6 buffer too short",
			b:           []byte{0x06, 0xf4, 0x27, 0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected:    0,
			expectedEp:  identity.Endpoint{},
			wantErr:     true,
			errContains: "buffer too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var ep identity.Endpoint

			got, err := ep.Parse(tt.b)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("error mismatch:\n\tgot:  %q\n\twant: %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !ep.IP.Equal(tt.expectedEp.IP) {
				t.Fatalf("IP mismatch:\n\tgot:  %v\n\twant: %v", ep.IP, tt.expectedEp.IP)
			}

			if ep.Port != tt.expectedEp.Port {
				t.Fatalf("Port mismatch:\n\tgot:  %v\n\twant: %v", ep.Port, tt.expectedEp.Port)
			}

			if got != tt.expected {
				t.Fatalf("return value mismatch:\n\tgot:  %v\n\twant: %v", got, tt.expected)
			}
		})
	}
}

func TestEndpoint_RoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []identity.Endpoint{
		{IP: net.ParseIP("8.8.8.8"), Port: 53},
		{IP: net.ParseIP("::1"), Port: 8080},
		{IP: net.ParseIP("192.168.1.1"), Port: 443},
		{IP: net.ParseIP("2001:db8::1"), Port: 80},
		{IP: net.ParseIP("127.0.0.1"), Port: 0},
		{IP: net.ParseIP("fe80::1"), Port: 65535},
		{IP: net.ParseIP("::ffff:192.0.2.1"), Port: 9000},
	}

	for _, original := range testCases {
		t.Run(original.String(), func(t *testing.T) {
			t.Parallel()

			bytes, err := original.Bytes()
			if err != nil {
				t.Fatalf("Bytes() failed: %v", err)
			}

			var parsed identity.Endpoint
			n, err := parsed.Parse(bytes)
			if err != nil {
				t.Fatalf("Parse() failed: %v", err)
			}

			if n != len(bytes) {
				t.Fatalf("didn't consume all bytes: consumed %d, total %d", n, len(bytes))
			}

			if !parsed.IP.Equal(original.IP) {
				t.Fatalf("IP mismatch:\n\toriginal: %v\n\tparsed:   %v", original.IP, parsed.IP)
			}

			if parsed.Port != original.Port {
				t.Fatalf("Port mismatch:\n\toriginal: %v\n\tparsed:   %v", original.Port, parsed.Port)
			}
		})
	}
}

func TestNewEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ipStr       string
		port        uint16
		expected    identity.Endpoint
		wantErr     bool
		errContains string
	}{
		{
			name:  "valid IPv4",
			ipStr: "192.168.1.1",
			port:  8080,
			expected: identity.Endpoint{
				IP:   net.ParseIP("192.168.1.1"),
				Port: 8080,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name:  "valid IPv6",
			ipStr: "2001:db8::1",
			port:  443,
			expected: identity.Endpoint{
				IP:   net.ParseIP("2001:db8::1"),
				Port: 443,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name:  "valid IPv6 loopback",
			ipStr: "::1",
			port:  9000,
			expected: identity.Endpoint{
				IP:   net.ParseIP("::1"),
				Port: 9000,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "invalid IP",
			ipStr:       "invalid",
			port:        8080,
			expected:    identity.Endpoint{},
			wantErr:     true,
			errContains: "invalid IP address",
		},
		{
			name:        "empty IP",
			ipStr:       "",
			port:        8080,
			expected:    identity.Endpoint{},
			wantErr:     true,
			errContains: "invalid IP address",
		},
		{
			name:        "port 0",
			ipStr:       "192.168.1.1",
			port:        0,
			expected:    identity.Endpoint{},
			wantErr:     true,
			errContains: "invalid port",
		},
		{
			name:  "port 65535",
			ipStr: "192.168.1.1",
			port:  65535,
			expected: identity.Endpoint{
				IP:   net.ParseIP("192.168.1.1"),
				Port: 65535,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name:  "port 1",
			ipStr: "127.0.0.1",
			port:  1,
			expected: identity.Endpoint{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 1,
			},
			wantErr:     false,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := identity.NewEndpoint(tt.ipStr, tt.port)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("error mismatch:\n\tgot:  %q\n\twant: %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !got.IP.Equal(tt.expected.IP) {
				t.Fatalf("IP mismatch:\n\tgot:  %v\n\twant: %v", got.IP, tt.expected.IP)
			}

			if got.Port != tt.expected.Port {
				t.Fatalf("Port mismatch:\n\tgot:  %v\n\twant: %v", got.Port, tt.expected.Port)
			}
		})
	}
}

func TestParseEpFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ep          string
		expected    identity.Endpoint
		wantErr     bool
		errContains string
	}{
		{
			name: "valid IPv4",
			ep:   "192.168.1.1:8080",
			expected: identity.Endpoint{
				IP:   net.ParseIP("192.168.1.1"),
				Port: 8080,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name: "valid IPv6",
			ep:   "[2001:db8::1]:443",
			expected: identity.Endpoint{
				IP:   net.ParseIP("2001:db8::1"),
				Port: 443,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name: "IPv6 loopback",
			ep:   "[::1]:9000",
			expected: identity.Endpoint{
				IP:   net.ParseIP("::1"),
				Port: 9000,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name: "IPv4 localhost",
			ep:   "127.0.0.1:80",
			expected: identity.Endpoint{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 80,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "missing port",
			ep:          "192.168.1.1",
			expected:    identity.Endpoint{},
			wantErr:     true,
			errContains: "invalid endpoint format",
		},
		{
			name:        "missing brackets for IPv6",
			ep:          "2001:db8::1:443",
			expected:    identity.Endpoint{},
			wantErr:     true,
			errContains: "invalid endpoint format",
		},
		{
			name:        "invalid IP",
			ep:          "invalid:8080",
			expected:    identity.Endpoint{},
			wantErr:     true,
			errContains: "invalid IP address",
		},
		{
			name:        "port 0",
			ep:          "192.168.1.1:0",
			expected:    identity.Endpoint{},
			wantErr:     true,
			errContains: "invalid port",
		},
		{
			name:        "negative port",
			ep:          "192.168.1.1:-1",
			expected:    identity.Endpoint{},
			wantErr:     true,
			errContains: "invalid port",
		},
		{
			name:        "port overflow",
			ep:          "192.168.1.1:65536",
			expected:    identity.Endpoint{},
			wantErr:     true,
			errContains: "invalid port",
		},
		{
			name:        "non-numeric port",
			ep:          "192.168.1.1:abc",
			expected:    identity.Endpoint{},
			wantErr:     true,
			errContains: "invalid port",
		},
		{
			name:        "empty string",
			ep:          "",
			expected:    identity.Endpoint{},
			wantErr:     true,
			errContains: "invalid endpoint format",
		},
		{
			name: "port 65535",
			ep:   "10.0.0.1:65535",
			expected: identity.Endpoint{
				IP:   net.ParseIP("10.0.0.1"),
				Port: 65535,
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name: "port 1",
			ep:   "172.16.0.1:1",
			expected: identity.Endpoint{
				IP:   net.ParseIP("172.16.0.1"),
				Port: 1,
			},
			wantErr:     false,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := identity.ParseEpFromString(tt.ep)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("error mismatch:\n\tgot:  %q\n\twant: %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !got.IP.Equal(tt.expected.IP) {
				t.Fatalf("IP mismatch:\n\tgot:  %v\n\twant: %v", got.IP, tt.expected.IP)
			}

			if got.Port != tt.expected.Port {
				t.Fatalf("Port mismatch:\n\tgot:  %v\n\twant: %v", got.Port, tt.expected.Port)
			}
		})
	}
}

func TestEndpoint_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ep       identity.Endpoint
		expected string
	}{
		{
			name:     "IPv4",
			ep:       identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 8080},
			expected: "192.168.1.1:8080",
		},
		{
			name:     "IPv6",
			ep:       identity.Endpoint{IP: net.ParseIP("2001:db8::1"), Port: 443},
			expected: "[2001:db8::1]:443",
		},
		{
			name:     "IPv4 localhost",
			ep:       identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 80},
			expected: "127.0.0.1:80",
		},
		{
			name:     "IPv6 loopback",
			ep:       identity.Endpoint{IP: net.ParseIP("::1"), Port: 9000},
			expected: "[::1]:9000",
		},
		{
			name:     "IPv4 with port 0",
			ep:       identity.Endpoint{IP: net.ParseIP("10.0.0.1"), Port: 0},
			expected: "10.0.0.1:0",
		},
		{
			name:     "IPv6 with port 65535",
			ep:       identity.Endpoint{IP: net.ParseIP("fe80::1"), Port: 65535},
			expected: "[fe80::1]:65535",
		},
		{
			name:     "IPv4-mapped IPv6",
			ep:       identity.Endpoint{IP: net.ParseIP("::ffff:192.0.2.1"), Port: 8080},
			expected: "192.0.2.1:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.ep.String()

			if got != tt.expected {
				t.Fatalf("mismatch:\n\tgot:  %q\n\twant: %q", got, tt.expected)
			}
		})
	}
}

func TestEndpoint_IsIPv4(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ep       identity.Endpoint
		expected bool
	}{
		{
			name:     "IPv4",
			ep:       identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 8080},
			expected: true,
		},
		{
			name:     "IPv6",
			ep:       identity.Endpoint{IP: net.ParseIP("2001:db8::1"), Port: 443},
			expected: false,
		},
		{
			name:     "IPv4 localhost",
			ep:       identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 80},
			expected: true,
		},
		{
			name:     "IPv6 loopback",
			ep:       identity.Endpoint{IP: net.ParseIP("::1"), Port: 9000},
			expected: false,
		},
		{
			name:     "IPv4-mapped IPv6",
			ep:       identity.Endpoint{IP: net.ParseIP("::ffff:192.0.2.1"), Port: 8080},
			expected: true,
		},
		{
			name:     "nil IP",
			ep:       identity.Endpoint{IP: nil, Port: 8080},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.ep.IsIPv4()

			if got != tt.expected {
				t.Fatalf("mismatch:\n\tgot:  %v\n\twant: %v", got, tt.expected)
			}
		})
	}
}

func TestEndpoint_IsIPv6(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ep       identity.Endpoint
		expected bool
	}{
		{
			name:     "IPv4",
			ep:       identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 8080},
			expected: false,
		},
		{
			name:     "IPv6",
			ep:       identity.Endpoint{IP: net.ParseIP("2001:db8::1"), Port: 443},
			expected: true,
		},
		{
			name:     "IPv4 localhost",
			ep:       identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 80},
			expected: false,
		},
		{
			name:     "IPv6 loopback",
			ep:       identity.Endpoint{IP: net.ParseIP("::1"), Port: 9000},
			expected: true,
		},
		{
			name:     "IPv4-mapped IPv6",
			ep:       identity.Endpoint{IP: net.ParseIP("::ffff:192.0.2.1"), Port: 8080},
			expected: false,
		},
		{
			name:     "nil IP",
			ep:       identity.Endpoint{IP: nil, Port: 8080},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.ep.IsIPv6()

			if got != tt.expected {
				t.Fatalf("mismatch:\n\tgot:  %v\n\twant: %v", got, tt.expected)
			}
		})
	}
}

func TestEndpoint_Network(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ep       identity.Endpoint
		expected string
	}{
		{
			name:     "IPv4",
			ep:       identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 8080},
			expected: "tcp4",
		},
		{
			name:     "IPv6",
			ep:       identity.Endpoint{IP: net.ParseIP("2001:db8::1"), Port: 443},
			expected: "tcp6",
		},
		{
			name:     "IPv4 localhost",
			ep:       identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 80},
			expected: "tcp4",
		},
		{
			name:     "IPv6 loopback",
			ep:       identity.Endpoint{IP: net.ParseIP("::1"), Port: 9000},
			expected: "tcp6",
		},
		{
			name:     "IPv4-mapped IPv6",
			ep:       identity.Endpoint{IP: net.ParseIP("::ffff:192.0.2.1"), Port: 8080},
			expected: "tcp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.ep.Network()

			if got != tt.expected {
				t.Fatalf("mismatch:\n\tgot:  %q\n\twant: %q", got, tt.expected)
			}
		})
	}
}
