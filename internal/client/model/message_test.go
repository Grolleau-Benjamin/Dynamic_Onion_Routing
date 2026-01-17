package model_test

import (
	"net"
	"reflect"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client/model"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
)

func TestMessage_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		m        model.Message
		expected string
	}{
		{
			name: "valid input datas",
			m: model.Message{
				Dest: identity.Endpoint{IP: net.ParseIP("::1"), Port: 8080},
				Path: []identity.CryptoGroup{{
					Group: identity.RelayGroup{
						Relays: []identity.Relay{
							{
								Ep:     identity.Endpoint{IP: net.ParseIP("::1"), Port: 62503},
								UUID:   [16]byte{0x17, 0x02, 0x20, 0x02},
								PubKey: [32]byte{0x03, 0x07, 0x19, 0x72},
							},
						},
					},
					CipherKey: [32]byte{0x25, 0x06},
					EPK:       [32]byte{0x20, 0x03},
				}},
				Payload: []byte{0x01, 0x02, 0x03, 0x04},
			},
			expected: "Message{\n\tdest=[::1]:8080 \n\tpath=[{group=[[::1]:62503] cipher=2506... epk=2003...}] \n\tpayload_len=4\n}",
		},
		{
			name: "empty path and payload",
			m: model.Message{
				Dest:    identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 80},
				Path:    nil,
				Payload: nil,
			},
			expected: "Message{\n\tdest=127.0.0.1:80 \n\tpath=[] \n\tpayload_len=0\n}",
		},
		{
			name: "multiple layers in path",
			m: model.Message{
				Dest: identity.Endpoint{IP: net.ParseIP("::1"), Port: 8080},
				Path: []identity.CryptoGroup{
					{
						Group: identity.RelayGroup{
							Relays: []identity.Relay{
								{
									Ep: identity.Endpoint{IP: net.ParseIP("::1"), Port: 6001},
								},
							},
						},
						CipherKey: [32]byte{0xAA},
						EPK:       [32]byte{0xBB},
					},
					{
						Group: identity.RelayGroup{
							Relays: []identity.Relay{
								{
									Ep: identity.Endpoint{IP: net.ParseIP("::1"), Port: 6002},
								},
							},
						},
						CipherKey: [32]byte{0xCC},
						EPK:       [32]byte{0xDD},
					},
				},
				Payload: []byte("ping"),
			},
			expected: "Message{\n\tdest=[::1]:8080 \n\tpath=[{group=[[::1]:6001] cipher=AA00... epk=BB00...} {group=[[::1]:6002] cipher=CC00... epk=DD00...}] \n\tpayload_len=4\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.m.String()
			if got != tt.expected {
				t.Fatalf("String() mismatch: \ngot: %s\nwant: %s", got, tt.expected)
			}
		})
	}
}

func TestMessage_BuildFromInputConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ic       client.InputConfig
		expected model.Message
		wantErr  bool
		err      error
	}{
		{
			name: "valid input data",
			ic: client.InputConfig{
				OnionPath: "[::1]:62503,[::1]:62504|192.168.58.221:10",
				Dest:      "8.8.8.8:63",
				Payload:   "Who are you? Google?",
			},
			expected: model.Message{
				Dest: identity.Endpoint{IP: net.ParseIP("8.8.8.8"), Port: 63},
				Path: []identity.CryptoGroup{
					{
						Group: identity.RelayGroup{
							Relays: []identity.Relay{
								{Ep: identity.Endpoint{IP: net.ParseIP("::1"), Port: 62503}},
								{Ep: identity.Endpoint{IP: net.ParseIP("::1"), Port: 62504}},
							},
						},
					},
					{
						Group: identity.RelayGroup{
							Relays: []identity.Relay{
								{Ep: identity.Endpoint{IP: net.ParseIP("192.168.58.221"), Port: 10}},
							},
						},
					},
				},
				Payload: []byte("Who are you? Google?"),
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "empty payload",
			ic: client.InputConfig{
				OnionPath: "[::1]:62503",
				Dest:      "127.0.0.1:8080",
				Payload:   "",
			},
			expected: model.Message{
				Dest: identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 8080},
				Path: []identity.CryptoGroup{
					{
						Group: identity.RelayGroup{
							Relays: []identity.Relay{
								{Ep: identity.Endpoint{IP: net.ParseIP("::1"), Port: 62503}},
							},
						},
					},
				},
				Payload: []byte{},
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "empty destination",
			ic: client.InputConfig{
				OnionPath: "[::1]:62503",
				Dest:      "",
				Payload:   "test payload",
			},
			expected: model.Message{},
			wantErr:  true,
			err:      nil,
		},
		{
			name: "empty onion path",
			ic: client.InputConfig{
				OnionPath: "",
				Dest:      "127.0.0.1:8080",
				Payload:   "test payload",
			},
			expected: model.Message{},
			wantErr:  true,
			err:      nil,
		},
		{
			name: "single relay in path",
			ic: client.InputConfig{
				OnionPath: "192.168.1.1:9000",
				Dest:      "[::1]:443",
				Payload:   "Hi there, some ben?",
			},
			expected: model.Message{
				Dest: identity.Endpoint{IP: net.ParseIP("::1"), Port: 443},
				Path: []identity.CryptoGroup{
					{
						Group: identity.RelayGroup{
							Relays: []identity.Relay{
								{Ep: identity.Endpoint{IP: net.ParseIP("192.168.1.1"), Port: 9000}},
							},
						},
					},
				},
				Payload: []byte("Hi there, some ben?"),
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "multiple groups with multiple relays",
			ic: client.InputConfig{
				OnionPath: "[::1]:1000,[::1]:1001,[::1]:1002|192.168.1.10:2000,192.168.1.11:2001|10.0.0.1:3000",
				Dest:      "8.8.4.4:53",
				Payload:   "A really huge path",
			},
			expected: model.Message{
				Dest: identity.Endpoint{IP: net.ParseIP("8.8.4.4"), Port: 53},
				Path: []identity.CryptoGroup{
					{
						Group: identity.RelayGroup{
							Relays: []identity.Relay{
								{Ep: identity.Endpoint{IP: net.ParseIP("::1"), Port: 1000}},
								{Ep: identity.Endpoint{IP: net.ParseIP("::1"), Port: 1001}},
								{Ep: identity.Endpoint{IP: net.ParseIP("::1"), Port: 1002}},
							},
						},
					},
					{
						Group: identity.RelayGroup{
							Relays: []identity.Relay{
								{Ep: identity.Endpoint{IP: net.ParseIP("192.168.1.10"), Port: 2000}},
								{Ep: identity.Endpoint{IP: net.ParseIP("192.168.1.11"), Port: 2001}},
							},
						},
					},
					{
						Group: identity.RelayGroup{
							Relays: []identity.Relay{
								{Ep: identity.Endpoint{IP: net.ParseIP("10.0.0.1"), Port: 3000}},
							},
						},
					},
				},
				Payload: []byte("A really huge path"),
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "all empty inputs",
			ic: client.InputConfig{
				OnionPath: "",
				Dest:      "",
				Payload:   "",
			},
			expected: model.Message{},
			wantErr:  true,
			err:      nil,
		},
		{
			name: "whitespace only inputs",
			ic: client.InputConfig{
				OnionPath: "   ",
				Dest:      "  ",
				Payload:   "   ",
			},
			expected: model.Message{},
			wantErr:  true,
			err:      nil,
		},
		{
			name: "large payload",
			ic: client.InputConfig{
				OnionPath: "[::1]:9000",
				Dest:      "127.0.0.1:80",
				Payload:   string(make([]byte, 10000)),
			},
			expected: model.Message{
				Dest: identity.Endpoint{IP: net.ParseIP("127.0.0.1"), Port: 80},
				Path: []identity.CryptoGroup{
					{
						Group: identity.RelayGroup{
							Relays: []identity.Relay{
								{Ep: identity.Endpoint{IP: net.ParseIP("::1"), Port: 9000}},
							},
						},
					},
				},
				Payload: make([]byte, 10000),
			},
			wantErr: false,
			err:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := model.BuildFromInputConfig(tt.ic)

			if !tt.wantErr && err != nil {
				t.Fatalf("BuildFromInputConfig() unexpected error:\n\tgot %v\n\texpected: %v", err, tt.err)
			}

			if tt.wantErr && err == nil {
				t.Fatalf("BuildFromInputConfig() expected error but got nil")
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("BuildFromInputConfig() mismatch:\n\tgot: %+v\n\twant: %+v", got, tt.expected)
			}
		})
	}
}
