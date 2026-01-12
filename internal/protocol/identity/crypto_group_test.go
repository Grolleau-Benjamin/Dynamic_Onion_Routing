package identity_test

import (
	"bytes"
	"net"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"golang.org/x/crypto/curve25519"
)

func TestCryptoGroup_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		group    identity.CryptoGroup
		expected string
	}{
		{
			name: "normal case",
			group: identity.CryptoGroup{
				Group: identity.RelayGroup{
					Relays: []identity.Relay{
						{
							Ep: identity.Endpoint{
								IP:   net.ParseIP("127.0.0.1"),
								Port: 62503,
							},
							UUID: [16]byte{
								0x01, 0x02, 0x03, 0x04,
							},
							PubKey: [32]byte{
								0x05, 0x06, 0x07, 0x08,
							},
						},
					},
				},
				CipherKey: [32]byte{
					0xaa, 0xbb, 0xcc, 0xdd,
				},
				EPK: [32]byte{
					0xee, 0xff, 0x11, 0x22,
				},
				ESK: [32]byte{
					0x33, 0x44, 0x55, 0x66,
				},
			},
			expected: "{group=[127.0.0.1:62503] cipher=AABB... epk=EEFF...}",
		},
		{
			name: "empty case",
			group: identity.CryptoGroup{
				Group:     identity.RelayGroup{},
				CipherKey: [32]byte{},
				EPK:       [32]byte{},
				ESK:       [32]byte{},
			},
			expected: "{group=[] cipher=0000... epk=0000...}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.group.String()
			if got != tt.expected {
				t.Fatalf("String() mismatch:\n got: %s\nwant: %s", got, tt.expected)
			}
		})
	}
}

func TestCryptoGroup_GenerateCryptoMaterial(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		group    identity.CryptoGroup
		wantErr  bool
		expected string
	}{
		{
			name: "normal case",
			group: identity.CryptoGroup{
				Group:     identity.RelayGroup{},
				CipherKey: [32]byte{},
				EPK:       [32]byte{},
				ESK:       [32]byte{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.group.GenerateCryptoMaterial()
			if (err != nil) && !tt.wantErr {
				t.Fatalf("GenerateCryptoMaterial() Unexpected error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if tt.group.CipherKey == [32]byte{} {
					t.Fatal("CipherKey was not generated")
				}
				if tt.group.EPK == [32]byte{} {
					t.Fatal("EPK was not generated")
				}
				if tt.group.ESK == [32]byte{} {
					t.Fatal("ESK was not generated")
				}
				derivedPK, err := curve25519.X25519(tt.group.ESK[:], curve25519.Basepoint)
				if err != nil || !bytes.Equal(derivedPK, tt.group.EPK[:]) {
					t.Fatal("Failed to derive expected public key from private key")
				}
			}
		})
	}
}
