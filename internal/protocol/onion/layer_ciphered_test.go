package onion_test

import (
	"bytes"
	"net"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/onion"
)

func TestBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		layer       onion.OnionLayerCiphered
		want        []byte
		wantErr     bool
		errContains string
	}{
		{
			name: "to much wrapped keys",
			layer: onion.OnionLayerCiphered{
				NextHops: make([]identity.Endpoint, onion.MaxWrappedKey+1),
			},
			wantErr:     true,
			errContains: "too much wrappedKeys",
		},
		{
			name: "normal case",
			layer: onion.OnionLayerCiphered{
				LastServer:        true,
				NextHops:          []identity.Endpoint{},
				UtilPayloadLength: 42,
				Payload:           []byte("DOR >>> TOR"),
			},
			want: []byte{
				0x08,
				0x00, 0x2A,
				'D', 'O', 'R', ' ', '>', '>', '>', ' ',
				'T', 'O', 'R',
			},
			wantErr:     false,
			errContains: "",
		},
		{
			name: "with next hops",
			layer: onion.OnionLayerCiphered{
				LastServer:        false,
				NextHops:          []identity.Endpoint{{IP: net.ParseIP("8.8.8.8"), Port: 31033}, {IP: net.ParseIP("8.8.4.4"), Port: 29103}},
				UtilPayloadLength: 5,
				Payload:           []byte("DORI!"),
			},
			want:        []byte{2, 0, 5, 4, 121, 57, 8, 8, 8, 8, 4, 113, 175, 8, 8, 4, 4, 68, 79, 82, 73, 33},
			wantErr:     false,
			errContains: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.layer.Bytes()
			if (err != nil) && !tt.wantErr {
				t.Fatalf("OnionLayerCiphered.Bytes() \n\tunexpected error = %v, \n\twantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("OnionLayerCiphered.Bytes() \n\tgot  %v, \n\twant %v", got, tt.want)
			}
			if tt.wantErr && err != nil && tt.errContains != "" && !bytes.Contains([]byte(err.Error()), []byte(tt.errContains)) {
				t.Fatalf("OnionLayerCiphered.Bytes() \n\terror = %v, \n\twantErrContains %v", err, tt.errContains)
				return
			}
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		data        []byte
		want        onion.OnionLayerCiphered
		errContains string
		wantErr     bool
	}{
		{
			name:        "buffer too short",
			data:        []byte{0x00},
			errContains: "buffer too short",
			wantErr:     true,
		},
		{
			name: "normal case",
			data: []byte{
				0x08,
				0x00, 0x2A,
				'D', 'O', 'R', ' ', '>', '>', '>', ' ',
				'T', 'O', 'R',
			},
			want: onion.OnionLayerCiphered{
				LastServer:        true,
				NextHops:          []identity.Endpoint{},
				UtilPayloadLength: 42,
				Payload:           []byte("DOR >>> TOR"),
			},
			wantErr: false,
		},
		{
			name: "with next hops",
			data: []byte{2, 0, 5, 4, 121, 57, 8, 8, 8, 8, 4, 113, 175, 8, 8, 4, 4, 68, 79, 82, 73, 33},
			want: onion.OnionLayerCiphered{
				LastServer:        false,
				NextHops:          []identity.Endpoint{{IP: net.ParseIP("8.8.8.8"), Port: 31033}, {IP: net.ParseIP("8.8.4.4"), Port: 29103}},
				UtilPayloadLength: 5,
				Payload:           []byte("DORI!"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var olc onion.OnionLayerCiphered
			err := olc.Parse(tt.data)
			if (err != nil) && !tt.wantErr {
				t.Fatalf("OnionLayerCiphered.Parse() \n\tunexpected error = %v, \n\twantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" && !bytes.Contains([]byte(err.Error()), []byte(tt.errContains)) {
				t.Fatalf("OnionLayerCiphered.Parse() \n\terror = %v, \n\twantErrContains %v", err, tt.errContains)
				return
			}
			if !tt.wantErr {
				if olc.LastServer != tt.want.LastServer {
					t.Fatalf("OnionLayerCiphered.Parse() LastServer \n\tgot  %v, \n\twant %v", olc.LastServer, tt.want.LastServer)
				}
				if olc.UtilPayloadLength != tt.want.UtilPayloadLength {
					t.Fatalf("OnionLayerCiphered.Parse() UtilPayloadLength \n\tgot  %v, \n\twant %v", olc.UtilPayloadLength, tt.want.UtilPayloadLength)
				}
				if !bytes.Equal(olc.Payload, tt.want.Payload) {
					t.Fatalf("OnionLayerCiphered.Parse() Payload \n\tgot  %v, \n\twant %v", olc.Payload, tt.want.Payload)
				}
				if len(olc.NextHops) != len(tt.want.NextHops) {
					t.Fatalf("OnionLayerCiphered.Parse() NextHops length \n\tgot  %v, \n\twant %v", len(olc.NextHops), len(tt.want.NextHops))
				} else {
					for i := range olc.NextHops {
						if !olc.NextHops[i].IP.Equal(tt.want.NextHops[i].IP) || olc.NextHops[i].Port != tt.want.NextHops[i].Port {
							t.Fatalf("OnionLayerCiphered.Parse() NextHops[%d] \n\tgot  %v, \n\twant %v", i, olc.NextHops[i], tt.want.NextHops[i])
						}
					}
				}
			}
		})
	}
}

func TestRoundedTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		layer onion.OnionLayerCiphered
	}{
		{
			name: "normal case",
			layer: onion.OnionLayerCiphered{
				LastServer:        true,
				NextHops:          []identity.Endpoint{{IP: net.ParseIP("8.8.8.8"), Port: 31033}, {IP: net.ParseIP("8.8.4.4"), Port: 29103}},
				UtilPayloadLength: 21,
				Payload:           []byte("J'aime le fromage de vache"),
			},
		},
		{
			name: "no next hops",
			layer: onion.OnionLayerCiphered{
				LastServer:        false,
				NextHops:          []identity.Endpoint{},
				UtilPayloadLength: 15,
				Payload:           []byte("Range ton casque!"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := tt.layer.Bytes()
			if err != nil {
				t.Fatalf("OnionLayerCiphered.Bytes() unexpected error: %v", err)
			}

			var olc onion.OnionLayerCiphered
			err = olc.Parse(data)
			if err != nil {
				t.Fatalf("OnionLayerCiphered.Parse() unexpected error: %v", err)
			}

			if olc.LastServer != tt.layer.LastServer {
				t.Fatalf("OnionLayerCiphered rounded trip LastServer \n\tgot  %v, \n\twant %v", olc.LastServer, tt.layer.LastServer)
			}
			if olc.UtilPayloadLength != tt.layer.UtilPayloadLength {
				t.Fatalf("OnionLayerCiphered rounded trip UtilPayloadLength \n\tgot  %v, \n\twant %v", olc.UtilPayloadLength, tt.layer.UtilPayloadLength)
			}
			if !bytes.Equal(olc.Payload, tt.layer.Payload) {
				t.Fatalf("OnionLayerCiphered rounded trip Payload \n\tgot  %v, \n\twant %v", olc.Payload, tt.layer.Payload)
			}
			if len(olc.NextHops) != len(tt.layer.NextHops) {
				t.Fatalf("OnionLayerCiphered rounded trip NextHops length \n\tgot  %v, \n\twant %v", len(olc.NextHops), len(tt.layer.NextHops))
			} else {
				for i := range olc.NextHops {
					if !olc.NextHops[i].IP.Equal(tt.layer.NextHops[i].IP) || olc.NextHops[i].Port != tt.layer.NextHops[i].Port {
						t.Fatalf("OnionLayerCiphered rounded trip NextHops[%d] \n\tgot  %v, \n\twant %v", i, olc.NextHops[i], tt.layer.NextHops[i])
					}
				}
			}
		})
	}
}
