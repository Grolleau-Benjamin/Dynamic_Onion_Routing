package identity_test

import (
	"net"
	"strings"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
)

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		group    identity.RelayGroup
		expected string
	}{
		{
			name: "single relay",
			group: identity.RelayGroup{
				Relays: []identity.Relay{
					{
						Ep: identity.Endpoint{
							IP:   net.ParseIP("127.0.0.1"),
							Port: 62503,
						},
					},
				},
			},
			expected: "[127.0.0.1:62503]",
		},
		{
			name: "multiple relays",
			group: identity.RelayGroup{
				Relays: []identity.Relay{
					{
						Ep: identity.Endpoint{
							IP:   net.ParseIP("127.0.0.1"),
							Port: 62503,
						},
					},
					{
						Ep: identity.Endpoint{
							IP:   net.ParseIP("192.168.1.1"),
							Port: 62504,
						},
					},
				},
			},
			expected: "[127.0.0.1:62503, 192.168.1.1:62504]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.group.String()
			if got != tt.expected {
				t.Fatalf("String() mismatch:\n\t got: %s\n\twant: %s", got, tt.expected)
			}
		})
	}
}

func TestParseRelayPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		raw         string
		expected    []identity.RelayGroup
		errContains string
		wantErr     bool
	}{
		{
			name: "valid single group with single relay",
			raw:  "127.0.0.1:62504",
			expected: []identity.RelayGroup{
				{
					Relays: []identity.Relay{
						{
							Ep: identity.Endpoint{
								IP:   net.ParseIP("127.0.0.1"),
								Port: 62504,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid multiple groups with multiple relays",
			raw:  "[::1]:62503,127.0.0.1:62504|[::1]:62505",
			expected: []identity.RelayGroup{
				{
					Relays: []identity.Relay{
						{
							Ep: identity.Endpoint{
								IP:   net.ParseIP("::1"),
								Port: 62503,
							},
						},
						{
							Ep: identity.Endpoint{
								IP:   net.ParseIP("127.0.0.1"),
								Port: 62504,
							},
						},
					},
				},
				{
					Relays: []identity.Relay{
						{
							Ep: identity.Endpoint{
								IP:   net.ParseIP("::1"),
								Port: 62505,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "empty relay path",
			raw:         "   ",
			expected:    nil,
			errContains: "empty relay path",
			wantErr:     true,
		},
		{
			name:        "too many jumps",
			raw:         "[::1]:3103|[::1]:3103|[::1]:3103|[::1]:3103|[::1]:3103|[::1]:3103|[::1]:3103|[::1]:3103|[::1]:3103|[::1]:3103|[::1]:3103",
			expected:    nil,
			errContains: "too many jump",
			wantErr:     true,
		},
		{
			name:        "empty relay group",
			raw:         "[::1]:3103|   |[::1]:3103",
			expected:    nil,
			errContains: "empty relay group at",
			wantErr:     true,
		},
		{
			name:        "too many nodes in a group",
			raw:         "[::1]:3103,[::1]:3103,[::1]:3103,[::1]:3103,[::1]:3103,[::1]:3103,[::1]:3103,[::1]:3103,[::1]:3103,[::1]:3103,[::1]:3103",
			expected:    nil,
			errContains: "too many node at group",
			wantErr:     true,
		},
		{
			name:        "empty relay in group",
			raw:         "[::1]:3103,   ,[::1]:3103",
			expected:    nil,
			errContains: "empty relay in group",
			wantErr:     true,
		},
		{
			name:        "invalid relay endpoint",
			raw:         "invalid_relay",
			expected:    nil,
			errContains: "invalid relay",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := identity.ParseRelayPath(tt.raw)
			if (err != nil) && !tt.wantErr {
				t.Fatalf("ParseRelayPath(): \n\tunexpected error = %v, \n\twantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("ParseRelayPath(): \n\tgot error = %v, \n\twant error containing %q", err, tt.errContains)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.expected) {
					t.Fatalf("ParseRelayPath(): \n\tgot length = %d, \n\twant %d", len(got), len(tt.expected))
					return
				}
				for i, group := range got {
					if len(group.Relays) != len(tt.expected[i].Relays) {
						t.Fatalf("ParseRelayPath(): group %d \n\tgot relays = %d, \n\twant %d", i, len(group.Relays), len(tt.expected[i].Relays))
						return
					}
					for j, relay := range group.Relays {
						expectedRelay := tt.expected[i].Relays[j]
						if !relay.Ep.IP.Equal(expectedRelay.Ep.IP) {
							t.Errorf("ParseRelayPath(): group %d relay %d \n\tgot IP = %v, \n\twant %v", i, j, relay.Ep.IP, expectedRelay.Ep.IP)
						}
						if relay.Ep.Port != expectedRelay.Ep.Port {
							t.Errorf("ParseRelayPath(): group %d relay %d \n\tgot Port = %d, \n\twant %d", i, j, relay.Ep.Port, expectedRelay.Ep.Port)
						}
					}
				}
			}
		})
	}
}
