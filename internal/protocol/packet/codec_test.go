package packet_test

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
)

func TestReadPacket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		raw         []byte
		wantType    uint8
		wantLen     int
		wantErr     bool
		errContains string
	}{
		{
			name:     "Basic GetIdentityRequest",
			raw:      []byte{packet.TypeGetIdentityRequest, 0x00, 0x00},
			wantType: packet.TypeGetIdentityRequest,
			wantLen:  0,
			wantErr:  false,
		},
		{
			name: "GetIdentityResponse with valid payload",
			raw: func() []byte {
				buf := make([]byte, 3+48)
				buf[0] = packet.TypeGetIdentityResponse
				binary.BigEndian.PutUint16(buf[1:3], 48)
				for i := 3; i < len(buf); i++ {
					buf[i] = byte(i)
				}
				return buf
			}(),
			wantType: packet.TypeGetIdentityResponse,
			wantLen:  48,
			wantErr:  false,
		},
		{
			name: "OnionPacket with correct 4096 payload",
			raw: func() []byte {
				payload := make([]byte, 4096)
				buf := make([]byte, 3+len(payload))
				buf[0] = packet.TypeOnionPacket
				binary.BigEndian.PutUint16(buf[1:3], uint16(len(payload)))
				copy(buf[3:], payload)
				return buf
			}(),
			wantType: packet.TypeOnionPacket,
			wantLen:  4096,
			wantErr:  false,
		},
		{
			name:        "header too short",
			raw:         []byte{packet.TypeGetIdentityRequest, 0x00},
			wantErr:     true,
			errContains: "",
		},
		{
			name:        "empty data",
			raw:         []byte{},
			wantErr:     true,
			errContains: "",
		},
		{
			name:        "unknown packet type",
			raw:         []byte{0xFF, 0x00, 0x00},
			wantErr:     true,
			errContains: "unknown packet type",
		},
		{
			name:        "invalid payload length for GetIdentityResponse",
			raw:         []byte{packet.TypeGetIdentityResponse, 0x00, 0x10},
			wantErr:     true,
			errContains: "invalid payload length",
		},
		{
			name: "OnionPacket with wrong length 100",
			raw: func() []byte {
				payload := make([]byte, 100)
				buf := make([]byte, 3+len(payload))
				buf[0] = packet.TypeOnionPacket
				binary.BigEndian.PutUint16(buf[1:3], uint16(len(payload)))
				copy(buf[3:], payload)
				return buf
			}(),
			wantErr:     true,
			errContains: "invalid payload length",
		},
		{
			name:        "OnionPacket with zero length",
			raw:         []byte{packet.TypeOnionPacket, 0x00, 0x00},
			wantErr:     true,
			errContains: "invalid payload length",
		},
		{
			name: "payload shorter than expected",
			raw: func() []byte {
				buf := make([]byte, 3+10)
				buf[0] = packet.TypeGetIdentityResponse
				binary.BigEndian.PutUint16(buf[1:3], 48)
				return buf
			}(),
			wantErr:     true,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := packet.ReadPacket(bytes.NewReader(tt.raw))

			if (err != nil) != tt.wantErr {
				t.Fatalf("ReadPacket() err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
					t.Fatalf("ReadPacket() error mismatch:\n\tgot:  %v\n\twant to contain: %s", err, tt.errContains)
				}
				return
			}

			if p.Type() != tt.wantType {
				t.Fatalf("Type() mismatch:\n\tgot:  0x%02x\n\twant: 0x%02x", p.Type(), tt.wantType)
			}

			wireLen := int(binary.BigEndian.Uint16(tt.raw[1:packet.HeaderSize]))
			if wireLen != tt.wantLen {
				t.Fatalf("wire length mismatch:\n\tgot:  %d\n\twant: %d", wireLen, tt.wantLen)
			}

			if exp, ok := p.ExpectedLen(); ok && exp != wireLen {
				t.Fatalf("ExpectedLen() mismatch:\n\tgot:  %d\n\twant: %d", exp, wireLen)
			}
		})
	}
}

func TestWritePacket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		packet      packet.Packet
		wantLen     int
		wantType    uint8
		wantErr     bool
		errContains string
	}{
		{
			name:     "GetIdentityRequest",
			packet:   &packet.GetIdentityRequest{},
			wantLen:  3,
			wantType: packet.TypeGetIdentityRequest,
			wantErr:  false,
		},
		{
			name: "GetIdentityResponse with data",
			packet: &packet.GetIdentityResponse{
				Ruuid:     [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
				PublicKey: [32]byte{0xaa, 0xbb, 0xcc, 0xdd},
			},
			wantLen:  3 + 48,
			wantType: packet.TypeGetIdentityResponse,
			wantErr:  false,
		},
		{
			name: "GetIdentityResponse with zero values",
			packet: &packet.GetIdentityResponse{
				Ruuid:     [16]byte{},
				PublicKey: [32]byte{},
			},
			wantLen:  3 + 48,
			wantType: packet.TypeGetIdentityResponse,
			wantErr:  false,
		},
		{
			name: "GetIdentityResponse with max values",
			packet: &packet.GetIdentityResponse{
				Ruuid:     [16]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				PublicKey: [32]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
			wantLen:  3 + 48,
			wantType: packet.TypeGetIdentityResponse,
			wantErr:  false,
		},
		{
			name:     "OnionPacket with 4096 bytes",
			packet:   &packet.OnionPacket{Data: [4096]byte{}},
			wantLen:  3 + 4096,
			wantType: packet.TypeOnionPacket,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := packet.WritePacket(&buf, tt.packet)

			if (err != nil) != tt.wantErr {
				t.Fatalf("WritePacket() err = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
					t.Fatalf("WritePacket() error mismatch:\n\tgot:  %v\n\twant to contain: %s", err, tt.errContains)
				}
				return
			}

			written := buf.Bytes()

			if len(written) != tt.wantLen {
				t.Fatalf("written length mismatch:\n\tgot:  %d\n\twant: %d", len(written), tt.wantLen)
			}

			if written[0] != tt.wantType {
				t.Fatalf("packet type mismatch:\n\tgot:  0x%02x\n\twant: 0x%02x", written[0], tt.wantType)
			}

			payloadLen := binary.BigEndian.Uint16(written[1:3])
			expectedPayloadLen := tt.wantLen - 3

			if int(payloadLen) != expectedPayloadLen {
				t.Fatalf("payload length mismatch:\n\tgot:  %d\n\twant: %d", payloadLen, expectedPayloadLen)
			}
		})
	}
}

func TestWritePacket_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		packet packet.Packet
	}{
		{
			name:   "GetIdentityRequest roundtrip",
			packet: &packet.GetIdentityRequest{},
		},
		{
			name: "GetIdentityResponse roundtrip",
			packet: &packet.GetIdentityResponse{
				Ruuid:     [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
				PublicKey: [32]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00},
			},
		},
		{
			name:   "OnionPacket roundtrip",
			packet: &packet.OnionPacket{Data: [4096]byte{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := packet.WritePacket(&buf, tt.packet)
			if err != nil {
				t.Fatalf("WritePacket() failed: %v", err)
			}

			p, err := packet.ReadPacket(&buf)
			if err != nil {
				t.Fatalf("ReadPacket() failed: %v", err)
			}

			if p.Type() != tt.packet.Type() {
				t.Errorf("roundtrip type mismatch:\n\tgot:  0x%02x\n\twant: 0x%02x", p.Type(), tt.packet.Type())
			}

			if resp, ok := tt.packet.(*packet.GetIdentityResponse); ok {
				readResp := p.(*packet.GetIdentityResponse)
				if readResp.Ruuid != resp.Ruuid {
					t.Errorf("Ruuid mismatch:\n\tgot:  %x\n\twant: %x", readResp.Ruuid, resp.Ruuid)
				}
				if readResp.PublicKey != resp.PublicKey {
					t.Errorf("PublicKey mismatch:\n\tgot:  %x\n\twant: %x", readResp.PublicKey, resp.PublicKey)
				}
			}
		})
	}
}
