package onion

import (
	"crypto/rand"
	"net"
	"strings"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"golang.org/x/crypto/curve25519"
)

func generateValidX25519Key() [32]byte {
	var priv [32]byte
	if _, err := rand.Read(priv[:]); err != nil {
		return [32]byte{}
	}
	pubSlice, err := curve25519.X25519(priv[:], curve25519.Basepoint)
	if err != nil {
		return [32]byte{}
	}
	var key [32]byte
	copy(key[:], pubSlice)
	return key
}

func TestComputePathOverhead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     []identity.CryptoGroup
		dest     identity.Endpoint
		expected int
	}{
		{
			name: "single hop path",
			path: []identity.CryptoGroup{
				{
					Group: identity.RelayGroup{
						Relays: []identity.Relay{
							{Ep: identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080}},
						},
					},
				},
			},
			dest:     identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000},
			expected: 311,
		},
		{
			name: "two hop path",
			path: []identity.CryptoGroup{
				{
					Group: identity.RelayGroup{
						Relays: []identity.Relay{
							{Ep: identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080}},
						},
					},
				},
				{
					Group: identity.RelayGroup{
						Relays: []identity.Relay{
							{Ep: identity.Endpoint{IP: net.ParseIP("127.0.0.2"), Port: 8081}},
						},
					},
				},
			},
			dest:     identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000},
			expected: 612,
		},
		{
			name: "multiple relays per group",
			path: []identity.CryptoGroup{
				{
					Group: identity.RelayGroup{
						Relays: []identity.Relay{
							{Ep: identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080}},
							{Ep: identity.Endpoint{IP: net.ParseIP("127.0.0.2"), Port: 8081}},
							{Ep: identity.Endpoint{IP: net.ParseIP("127.0.0.3"), Port: 8082}},
						},
					},
				},
			},
			dest:     identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000},
			expected: 325,
		},
		{
			name: "max hops path",
			path: func() []identity.CryptoGroup {
				groups := make([]identity.CryptoGroup, MaxJump)
				for i := range groups {
					groups[i] = identity.CryptoGroup{
						Group: identity.RelayGroup{
							Relays: []identity.Relay{
								{Ep: identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080}},
							},
						},
					}
				}
				return groups
			}(),
			dest:     identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000},
			expected: 1515,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := computePathOverhead(tt.path, tt.dest)
			if got != tt.expected {
				t.Fatalf("computePathOverhead() unexpected value\n\tgot: %d\n\twant: %d", got, tt.expected)
			}
		})
	}
}

func TestBuildOnion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		dest        identity.Endpoint
		path        []identity.CryptoGroup
		payload     []byte
		wantErr     bool
		errContains string
	}{
		{
			name: "valid single hop",
			dest: identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000},
			path: []identity.CryptoGroup{
				{
					Group: identity.RelayGroup{
						Relays: []identity.Relay{
							{
								Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
								PubKey: generateValidX25519Key(),
							},
						},
					},
					EPK: generateValidX25519Key(),
				},
			},
			payload: []byte("test payload"),
			wantErr: false,
		},
		{
			name: "valid two hops",
			dest: identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000},
			path: []identity.CryptoGroup{
				{
					Group: identity.RelayGroup{
						Relays: []identity.Relay{
							{
								Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
								PubKey: generateValidX25519Key(),
							},
						},
					},
					EPK: generateValidX25519Key(),
				},
				{
					Group: identity.RelayGroup{
						Relays: []identity.Relay{
							{
								Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
								PubKey: generateValidX25519Key(),
							},
						},
					},
					EPK: generateValidX25519Key(),
				},
			},
			payload: []byte("test payload for two hops"),
			wantErr: false,
		},
		{
			name:        "empty path",
			dest:        identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000},
			path:        []identity.CryptoGroup{},
			payload:     []byte("test"),
			wantErr:     true,
			errContains: "path cannot be empty",
		},
		{
			name:        "nil path",
			dest:        identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000},
			path:        nil,
			payload:     []byte("test"),
			wantErr:     true,
			errContains: "path cannot be empty",
		},
		{
			name: "too many hops",
			dest: identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000},
			path: func() []identity.CryptoGroup {
				groups := make([]identity.CryptoGroup, MaxJump+1)
				for i := range groups {
					groups[i] = identity.CryptoGroup{
						Group: identity.RelayGroup{
							Relays: []identity.Relay{
								{
									Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
									PubKey: generateValidX25519Key(),
								},
							},
						},
						EPK: generateValidX25519Key(),
					}
				}
				return groups
			}(),
			payload:     []byte("test"),
			wantErr:     true,
			errContains: "max jump",
		},
		{
			name: "payload too large",
			dest: identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000},
			path: []identity.CryptoGroup{
				{
					Group: identity.RelayGroup{
						Relays: []identity.Relay{
							{
								Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
								PubKey: generateValidX25519Key(),
							},
						},
					},
					EPK: generateValidX25519Key(),
				},
			},
			payload:     make([]byte, PacketSize),
			wantErr:     true,
			errContains: "payload too large",
		},
		{
			name: "empty payload",
			dest: identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000},
			path: []identity.CryptoGroup{
				{
					Group: identity.RelayGroup{
						Relays: []identity.Relay{
							{
								Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
								PubKey: generateValidX25519Key(),
							},
						},
					},
					EPK: generateValidX25519Key(),
				},
			},
			payload: []byte{},
			wantErr: false,
		},
		{
			name: "max hops with small payload",
			dest: identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000},
			path: func() []identity.CryptoGroup {
				groups := make([]identity.CryptoGroup, MaxJump)
				for i := range groups {
					groups[i] = identity.CryptoGroup{
						Group: identity.RelayGroup{
							Relays: []identity.Relay{
								{
									Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
									PubKey: generateValidX25519Key(),
								},
							},
						},
						EPK: generateValidX25519Key(),
					}
				}
				return groups
			}(),
			payload: []byte("small"),
			wantErr: false,
		},
		{
			name: "multiple relays in group",
			dest: identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000},
			path: []identity.CryptoGroup{
				{
					Group: identity.RelayGroup{
						Relays: []identity.Relay{
							{
								Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
								PubKey: generateValidX25519Key(),
							},
							{
								Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
								PubKey: generateValidX25519Key(),
							},
							{
								Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
								PubKey: generateValidX25519Key(),
							},
						},
					},
					EPK: generateValidX25519Key(),
				},
			},
			payload: []byte("payload with multiple relays"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			layer, err := BuildOnion(tt.dest, tt.path, tt.payload)

			if (err != nil) && !tt.wantErr {
				t.Fatalf("BuildOnion() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
					t.Fatalf("error mismatch:\n\tgot:  %v\n\twant to contain: %s", err, tt.errContains)
				}
				return
			}

			if layer == nil {
				t.Fatal("expected non-nil layer")
			}

			if layer.EPK != tt.path[0].EPK {
				t.Fatalf("EPK mismatch:\n\tgot:  %x\n\twant: %x", layer.EPK, tt.path[0].EPK)
			}

			if len(layer.CipherText) == 0 {
				t.Error("expected non-empty CipherText")
			}
		})
	}
}

func TestBuildOnion_OutputSize(t *testing.T) {
	t.Parallel()

	dest := identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000}
	path := []identity.CryptoGroup{
		{
			Group: identity.RelayGroup{
				Relays: []identity.Relay{
					{
						Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
						PubKey: generateValidX25519Key(),
					},
				},
			},
			EPK:       generateValidX25519Key(),
			CipherKey: [32]byte{0xbb},
		},
	}
	payload := []byte("test payload")

	layer, err := BuildOnion(dest, path, payload)
	if err != nil {
		t.Fatalf("BuildOnion() failed: %v", err)
	}

	layerBytes, err := layer.BytesPadded()
	if err != nil {
		t.Fatalf("BytesPadded() failed: %v", err)
	}

	if len(layerBytes) != PacketSize {
		t.Fatalf("output size mismatch:\n\tgot:  %d\n\twant: %d",
			len(layerBytes), PacketSize)
	}
}

func BenchmarkBuildOnion(b *testing.B) {
	dest := identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000}
	path := []identity.CryptoGroup{
		{
			Group: identity.RelayGroup{
				Relays: []identity.Relay{
					{
						Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
						PubKey: generateValidX25519Key(),
					},
				},
			},
			EPK:       generateValidX25519Key(),
			ESK:       generateValidX25519Key(),
			CipherKey: generateValidX25519Key(),
		},
		{
			Group: identity.RelayGroup{
				Relays: []identity.Relay{
					{
						Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.2"), Port: 8081},
						PubKey: generateValidX25519Key(),
					},
				},
			},
			EPK:       generateValidX25519Key(),
			ESK:       generateValidX25519Key(),
			CipherKey: generateValidX25519Key(),
		},
		{
			Group: identity.RelayGroup{
				Relays: []identity.Relay{
					{
						Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.3"), Port: 8082},
						PubKey: generateValidX25519Key(),
					},
				},
			},
			EPK:       generateValidX25519Key(),
			ESK:       generateValidX25519Key(),
			CipherKey: generateValidX25519Key(),
		},
	}
	payload := []byte("test payload for benchmarking")

	b.ResetTimer()
	for b.Loop() {
		_, err := BuildOnion(dest, path, payload)
		if err != nil {
			b.Fatalf("BuildOnion() error = %v", err)
		}
	}
}

func BenchmarkBuildOnion1Hop(b *testing.B) {
	dest := identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000}
	path := []identity.CryptoGroup{
		{
			Group: identity.RelayGroup{
				Relays: []identity.Relay{
					{
						Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
						PubKey: generateValidX25519Key(),
					},
				},
			},
			EPK:       generateValidX25519Key(),
			ESK:       generateValidX25519Key(),
			CipherKey: generateValidX25519Key(),
		},
	}
	payload := []byte("test payload")

	b.ResetTimer()
	for b.Loop() {
		_, err := BuildOnion(dest, path, payload)
		if err != nil {
			b.Fatalf("BuildOnion() error = %v", err)
		}
	}
}

func BenchmarkBuildOnion5Hops(b *testing.B) {
	dest := identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000}
	path := make([]identity.CryptoGroup, MaxJump)
	for i := range path {
		path[i] = identity.CryptoGroup{
			Group: identity.RelayGroup{
				Relays: []identity.Relay{
					{
						Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: uint16(8080 + i)},
						PubKey: generateValidX25519Key(),
					},
				},
			},
			EPK:       generateValidX25519Key(),
			ESK:       generateValidX25519Key(),
			CipherKey: generateValidX25519Key(),
		}
	}
	payload := []byte("test payload")

	b.ResetTimer()
	for b.Loop() {
		_, err := BuildOnion(dest, path, payload)
		if err != nil {
			b.Fatalf("BuildOnion() error = %v", err)
		}
	}
}

func BenchmarkBuildOnionLargePayload(b *testing.B) {
	dest := identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000}
	path := []identity.CryptoGroup{
		{
			Group: identity.RelayGroup{
				Relays: []identity.Relay{
					{
						Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
						PubKey: generateValidX25519Key(),
					},
				},
			},
			EPK:       generateValidX25519Key(),
			ESK:       generateValidX25519Key(),
			CipherKey: generateValidX25519Key(),
		},
	}
	payload := make([]byte, 2048)

	b.ResetTimer()
	for b.Loop() {
		_, err := BuildOnion(dest, path, payload)
		if err != nil {
			b.Fatalf("BuildOnion() error = %v", err)
		}
	}
}

func BenchmarkBuildOnionMultipleRelaysPerGroup(b *testing.B) {
	dest := identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000}
	path := []identity.CryptoGroup{
		{
			Group: identity.RelayGroup{
				Relays: []identity.Relay{
					{
						Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
						PubKey: generateValidX25519Key(),
					},
					{
						Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.2"), Port: 8081},
						PubKey: generateValidX25519Key(),
					},
					{
						Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.3"), Port: 8082},
						PubKey: generateValidX25519Key(),
					},
				},
			},
			EPK:       generateValidX25519Key(),
			ESK:       generateValidX25519Key(),
			CipherKey: generateValidX25519Key(),
		},
	}
	payload := []byte("test payload")

	b.ResetTimer()
	for b.Loop() {
		_, err := BuildOnion(dest, path, payload)
		if err != nil {
			b.Fatalf("BuildOnion() error = %v", err)
		}
	}
}

func BenchmarkBuildOnion3Hops3Relays(b *testing.B) {
	dest := identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000}
	path := make([]identity.CryptoGroup, 3)
	for i := range path {
		path[i] = identity.CryptoGroup{
			Group: identity.RelayGroup{
				Relays: []identity.Relay{
					{
						Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: uint16(8080 + i*10)},
						UUID:   [16]byte{byte(i*3 + 1)},
						PubKey: generateValidX25519Key(),
					},
					{
						Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.2"), Port: uint16(8081 + i*10)},
						UUID:   [16]byte{byte(i*3 + 2)},
						PubKey: generateValidX25519Key(),
					},
					{
						Ep:     identity.Endpoint{IP: net.ParseIP("127.0.0.3"), Port: uint16(8082 + i*10)},
						UUID:   [16]byte{byte(i*3 + 3)},
						PubKey: generateValidX25519Key(),
					},
				},
			},
			EPK:       generateValidX25519Key(),
			ESK:       generateValidX25519Key(),
			CipherKey: generateValidX25519Key(),
		}
	}
	payload := []byte("test payload for full network simulation")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := BuildOnion(dest, path, payload)
		if err != nil {
			b.Fatalf("BuildOnion() error = %v", err)
		}
	}
}
