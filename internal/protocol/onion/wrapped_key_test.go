package onion_test

import (
	"bytes"
	"net"
	"strings"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/crypto"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/onion"
	"golang.org/x/crypto/curve25519"
)

func TestNewWrappedKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cryptoGrp   *identity.CryptoGroup
		errContains string
		wantErr     bool
	}{
		{
			name: "Valid group with 3 relays",
			cryptoGrp: &identity.CryptoGroup{
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
						{
							Ep: identity.Endpoint{
								IP:   net.ParseIP("127.0.0.2"),
								Port: 62504,
							},
							UUID: [16]byte{
								0x11, 0x12, 0x13, 0x14,
							},
							PubKey: [32]byte{
								0x15, 0x16, 0x17, 0x18,
							},
						},
						{
							Ep: identity.Endpoint{
								IP:   net.ParseIP("127.0.0.3"),
								Port: 62505,
							},
							UUID: [16]byte{
								0x21, 0x22, 0x23, 0x24,
							},
							PubKey: [32]byte{
								0x25, 0x26, 0x27, 0x28,
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
			wantErr:     false,
			errContains: "",
		},
		{
			name: "Valid group with 1 relay",
			cryptoGrp: &identity.CryptoGroup{
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
					0x11, 0x22, 0xcc, 0xdd,
				},
				EPK: [32]byte{
					0xee, 0xff, 0x11, 0x22,
				},
				ESK: [32]byte{
					0x33, 0x44, 0xaa, 0xbb,
				},
			},
			wantErr: false,
		},
		{
			name: "Valid group with 2 relays",
			cryptoGrp: &identity.CryptoGroup{
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
						{
							Ep: identity.Endpoint{
								IP:   net.ParseIP("127.0.0.2"),
								Port: 62504,
							},
							UUID: [16]byte{
								0x11, 0x12, 0x13, 0x14,
							},
							PubKey: [32]byte{
								0x15, 0x16, 0x17, 0x18,
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
			wantErr:     false,
			errContains: "",
		},
		{
			name: "Too many relays in group",
			cryptoGrp: &identity.CryptoGroup{
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
						{
							Ep: identity.Endpoint{
								IP:   net.ParseIP("127.0.0.2"),
								Port: 62504,
							},
							UUID: [16]byte{
								0x11, 0x12, 0x13, 0x14,
							},
							PubKey: [32]byte{
								0x15, 0x16, 0x17, 0x18,
							},
						},
						{
							Ep: identity.Endpoint{
								IP:   net.ParseIP("127.0.0.3"),
								Port: 62505,
							},
							UUID: [16]byte{
								0x21, 0x22, 0x23, 0x24,
							},
							PubKey: [32]byte{
								0x25, 0x26, 0x27, 0x28,
							},
						},
						{
							Ep: identity.Endpoint{
								IP:   net.ParseIP("127.0.0.4"),
								Port: 62506,
							},
							UUID: [16]byte{
								0x31, 0x32, 0x33, 0x34,
							},
							PubKey: [32]byte{
								0x35, 0x36, 0x37, 0x38,
							},
						},
						{
							Ep: identity.Endpoint{
								IP:   net.ParseIP("127.0.0.5"),
								Port: 62507,
							},
							UUID: [16]byte{
								0x41, 0x42, 0x43, 0x44,
							},
							PubKey: [32]byte{
								0x45, 0x46, 0x47, 0x48,
							},
						},
					},
				},
				CipherKey: [32]byte{
					0x11, 0x22, 0xcc, 0xdd,
				},
				EPK: [32]byte{
					0xee, 0xff, 0x11, 0x22,
				},
				ESK: [32]byte{
					0x33, 0x44, 0xaa, 0xbb,
				},
			},
			wantErr:     true,
			errContains: "too many relays in group:",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			finalKeys, err := onion.NewWrappedKeys(tt.cryptoGrp)

			if (err != nil) && !tt.wantErr {
				t.Fatalf("NewWrappedKeys() \n\tunexpected error = %v, \n\twantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("NewWrappedKeys() \n\terror = %v, \n\twantErrContains %v", err, tt.errContains)
			}

			count := 0

			for _, wk := range finalKeys {
				for _, relay := range tt.cryptoGrp.Group.Relays {

					sharedSecret, err := curve25519.X25519(tt.cryptoGrp.ESK[:], relay.PubKey[:])
					if err != nil {
						t.Fatalf("shared secret failed: %v", err)
					}

					wrappingKeySlice, err := crypto.HKDFSha256(
						sharedSecret,
						onion.HKDFSaltWrappedKey,
						onion.HKDFInfoWrappedKey,
					)
					if err != nil {
						t.Fatalf("HKDF failed: %v", err)
					}

					var wrappingKey [32]byte
					copy(wrappingKey[:], wrappingKeySlice)

					plaintext, err := crypto.ChachaDecrypt(
						wrappingKey,
						wk.Nonce,
						wk.CipherText[:],
						[]byte("DORv1:WrappedKey"),
					)
					if err != nil {
						continue
					}

					if len(plaintext) != 48 {
						t.Fatalf("invalid plaintext size: %d", len(plaintext))
					}

					uuid := plaintext[:16]
					cipherKey := plaintext[16:48]

					if !bytes.Equal(uuid, relay.UUID[:]) {
						t.Fatalf("UUID mismatch after decrypt")
					}
					if !bytes.Equal(cipherKey, tt.cryptoGrp.CipherKey[:]) {
						t.Fatalf("CipherKey mismatch after decrypt")
					}

					count++
					break
				}
			}

			if count != len(tt.cryptoGrp.Group.Relays) && !tt.wantErr {
				t.Fatalf("expected to recover %d wrapped keys, got %d", len(tt.cryptoGrp.Group.Relays), count)
			}
		})
	}
}
