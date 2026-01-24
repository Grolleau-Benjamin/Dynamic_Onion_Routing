package packet_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
)

func TestGetIdentityResponse_Type(t *testing.T) {
	t.Parallel()

	pkt := &packet.GetIdentityResponse{}
	got := pkt.Type()

	if got != packet.TypeGetIdentityResponse {
		t.Errorf("Type() mismatch:\n\tgot:  0x%02x\n\twant: 0x%02x", got, packet.TypeGetIdentityResponse)
	}
}

func TestGetIdentityResponse_ExpectedLen(t *testing.T) {
	t.Parallel()

	pkt := &packet.GetIdentityResponse{}
	length, ok := pkt.ExpectedLen()

	if !ok {
		t.Error("ExpectedLen() should return true for fixed length")
	}

	if length != 48 {
		t.Errorf("ExpectedLen() mismatch:\n\tgot:  %d\n\twant: 48", length)
	}
}

func TestGetIdentityResponse_Encode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		pkt         packet.GetIdentityResponse
		wantLen     int
		checkRuuid  bool
		checkPubKey bool
	}{
		{
			name: "encode with data",
			pkt: packet.GetIdentityResponse{
				Ruuid:     [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
				PublicKey: [32]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00},
			},
			wantLen:     48,
			checkRuuid:  true,
			checkPubKey: true,
		},
		{
			name: "encode with zero values",
			pkt: packet.GetIdentityResponse{
				Ruuid:     [16]byte{},
				PublicKey: [32]byte{},
			},
			wantLen:     48,
			checkRuuid:  true,
			checkPubKey: true,
		},
		{
			name: "encode with max values",
			pkt: packet.GetIdentityResponse{
				Ruuid:     [16]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				PublicKey: [32]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
			wantLen:     48,
			checkRuuid:  true,
			checkPubKey: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := tt.pkt.Encode(&buf)
			if err != nil {
				t.Fatalf("Encode() unexpected error: %v", err)
			}

			got := buf.Bytes()

			if len(got) != tt.wantLen {
				t.Errorf("encoded length mismatch:\n\tgot:  %d\n\twant: %d", len(got), tt.wantLen)
			}

			if tt.checkRuuid && !bytes.Equal(got[:16], tt.pkt.Ruuid[:]) {
				t.Errorf("Ruuid mismatch:\n\tgot:  %x\n\twant: %x", got[:16], tt.pkt.Ruuid[:])
			}

			if tt.checkPubKey && !bytes.Equal(got[16:48], tt.pkt.PublicKey[:]) {
				t.Errorf("PublicKey mismatch:\n\tgot:  %x\n\twant: %x", got[16:48], tt.pkt.PublicKey[:])
			}
		})
	}
}

func TestGetIdentityResponse_Decode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		data        []byte
		wantRuuid   [16]byte
		wantPubKey  [32]byte
		wantErr     bool
		errContains string
	}{
		{
			name: "decode valid data",
			data: func() []byte {
				buf := make([]byte, 48)
				copy(buf[:16], []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10})
				copy(buf[16:48], []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00})
				return buf
			}(),
			wantRuuid:  [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
			wantPubKey: [32]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00},
			wantErr:    false,
		},
		{
			name:       "decode all zeros",
			data:       make([]byte, 48),
			wantRuuid:  [16]byte{},
			wantPubKey: [32]byte{},
			wantErr:    false,
		},
		{
			name: "decode all max values",
			data: func() []byte {
				buf := make([]byte, 48)
				for i := range buf {
					buf[i] = 0xff
				}
				return buf
			}(),
			wantRuuid:  [16]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			wantPubKey: [32]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			wantErr:    false,
		},
		{
			name:        "decode data too short (only ruuid)",
			data:        make([]byte, 16),
			wantErr:     true,
			errContains: "",
		},
		{
			name:        "decode data too short (partial)",
			data:        make([]byte, 30),
			wantErr:     true,
			errContains: "",
		},
		{
			name:        "decode empty data",
			data:        []byte{},
			wantErr:     true,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var pkt packet.GetIdentityResponse
			err := pkt.Decode(bytes.NewReader(tt.data))

			if (err != nil) != tt.wantErr {
				t.Fatalf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
					t.Errorf("error mismatch:\n\tgot:  %v\n\twant to contain: %s", err, tt.errContains)
				}
				return
			}

			if pkt.Ruuid != tt.wantRuuid {
				t.Errorf("Ruuid mismatch:\n\tgot:  %x\n\twant: %x", pkt.Ruuid, tt.wantRuuid)
			}

			if pkt.PublicKey != tt.wantPubKey {
				t.Errorf("PublicKey mismatch:\n\tgot:  %x\n\twant: %x", pkt.PublicKey, tt.wantPubKey)
			}
		})
	}
}

func TestGetIdentityResponse_EncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pkt  packet.GetIdentityResponse
	}{
		{
			name: "roundtrip with data",
			pkt: packet.GetIdentityResponse{
				Ruuid:     [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
				PublicKey: [32]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00},
			},
		},
		{
			name: "roundtrip with zeros",
			pkt: packet.GetIdentityResponse{
				Ruuid:     [16]byte{},
				PublicKey: [32]byte{},
			},
		},
		{
			name: "roundtrip with max values",
			pkt: packet.GetIdentityResponse{
				Ruuid:     [16]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				PublicKey: [32]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := tt.pkt.Encode(&buf)
			if err != nil {
				t.Fatalf("Encode() failed: %v", err)
			}

			var decoded packet.GetIdentityResponse
			err = decoded.Decode(&buf)
			if err != nil {
				t.Fatalf("Decode() failed: %v", err)
			}

			if decoded.Ruuid != tt.pkt.Ruuid {
				t.Errorf("Ruuid mismatch after roundtrip:\n\tgot:  %x\n\twant: %x", decoded.Ruuid, tt.pkt.Ruuid)
			}

			if decoded.PublicKey != tt.pkt.PublicKey {
				t.Errorf("PublicKey mismatch after roundtrip:\n\tgot:  %x\n\twant: %x", decoded.PublicKey, tt.pkt.PublicKey)
			}
		})
	}
}
