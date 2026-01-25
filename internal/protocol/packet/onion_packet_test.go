package packet_test

import (
	"bytes"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/onion"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
)

func TestOnionPacket_Type(t *testing.T) {
	t.Parallel()

	p := &packet.OnionPacket{}
	if got := p.Type(); got != packet.TypeOnionPacket {
		t.Fatalf("Type() mismatch:\n\tgot:  0x%02x\n\twant: 0x%02x", got, packet.TypeOnionPacket)
	}
}

func TestOnionPacket_ExpectedLen(t *testing.T) {
	t.Parallel()

	p := &packet.OnionPacket{}
	gotLen, ok := p.ExpectedLen()

	if !ok {
		t.Fatalf("ExpectedLen() ok mismatch:\n\tgot:  %v\n\twant: true", ok)
	}
	if gotLen != onion.PacketSize {
		t.Fatalf("ExpectedLen() len mismatch:\n\tgot:  %d\n\twant: %d", gotLen, onion.PacketSize)
	}
}

func TestOnionPacket_Encode(t *testing.T) {
	t.Parallel()

	p := &packet.OnionPacket{}
	var buf bytes.Buffer

	if err := p.Encode(&buf); err != nil {
		t.Fatalf("Encode() unexpected error: %v", err)
	}
	if buf.Len() != onion.PacketSize {
		t.Fatalf("Encode() wrote wrong size:\n\tgot:  %d bytes\n\twant: %d bytes", buf.Len(), onion.PacketSize)
	}
}

func TestOnionPacket_Decode_ValidData(t *testing.T) {
	t.Parallel()

	p := &packet.OnionPacket{}
	input := make([]byte, onion.PacketSize)

	if err := p.Decode(bytes.NewReader(input)); err != nil {
		t.Fatalf("Decode() unexpected error: %v", err)
	}
}

func TestOnionPacket_Decode_InsufficientData(t *testing.T) {
	t.Parallel()

	p := &packet.OnionPacket{}
	input := make([]byte, 0)

	if err := p.Decode(bytes.NewReader(input)); err == nil {
		t.Fatalf("Decode() expected error for empty input, got nil")
	}
}

func TestOnionPacket_Roundtrip(t *testing.T) {
	t.Parallel()

	original := &packet.OnionPacket{}
	var buf bytes.Buffer

	if err := original.Encode(&buf); err != nil {
		t.Fatalf("Encode() unexpected error: %v", err)
	}

	decoded := &packet.OnionPacket{}
	if err := decoded.Decode(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("Decode() unexpected error: %v", err)
	}

	var buf2 bytes.Buffer
	if err := decoded.Encode(&buf2); err != nil {
		t.Fatalf("Encode() second pass unexpected error: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), buf2.Bytes()) {
		t.Fatalf("Roundtrip mismatch:\n\toriginal:  %v\n\troundtrip: %v", buf.Bytes(), buf2.Bytes())
	}
}
