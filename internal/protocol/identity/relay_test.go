package identity_test

import (
	"net"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
)

func TestRelay_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		relay    identity.Relay
		expected string
	}{
		{
			name: "Valid data",
			relay: identity.Relay{
				Ep:   identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 62503},
				UUID: [16]byte{0x01, 0x02, 0x03, 0x25, 0x06, 0x20, 0x03, 0x04},
				PubKey: [32]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
					0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
			},
			expected: "{ep=127.0.0.1:62503 uuid=01020325 pub=00010203}",
		},
		{
			name: "zero values",
			relay: identity.Relay{
				Ep:     identity.Endpoint{IP: net.ParseIP("0.0.0.0"), Port: 0},
				UUID:   [16]byte{},
				PubKey: [32]byte{},
			},
			expected: "{ep=0.0.0.0:0 uuid=00000000 pub=00000000}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.relay.String()
			if got != tt.expected {
				t.Fatalf("String() mismatch: \ngot: %s\nwant: %s", got, tt.expected)
			}
		})
	}
}

func TestRelay_HydrateIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		initialRelay identity.Relay
		uuid         [16]byte
		pubKey       [32]byte
	}{
		{
			name: "hydrate with valid identity",
			initialRelay: identity.Relay{
				Ep: identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 62503},
			},
			uuid:   [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
			pubKey: [32]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00},
		},
		{
			name: "hydrate overwrites existing identity",
			initialRelay: identity.Relay{
				Ep:     identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 8080},
				UUID:   [16]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				PubKey: [32]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
			uuid:   [16]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			pubKey: [32]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "hydrate with zero values",
			initialRelay: identity.Relay{
				Ep: identity.Endpoint{IP: net.ParseIP("10.0.0.1"), Port: 9000},
			},
			uuid:   [16]byte{},
			pubKey: [32]byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			relay := tt.initialRelay
			relay.HydrateIdentity(tt.uuid, tt.pubKey)

			if relay.UUID != tt.uuid {
				t.Errorf("UUID mismatch:\n  got:  %x\n  want: %x", relay.UUID, tt.uuid)
			}

			if relay.PubKey != tt.pubKey {
				t.Errorf("PubKey mismatch:\n  got:  %x\n  want: %x", relay.PubKey, tt.pubKey)
			}
		})
	}
}
