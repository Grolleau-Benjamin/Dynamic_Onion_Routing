package packet_test

import (
	"bytes"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
)

func TestGetIdentityRequest_Type(t *testing.T) {
	t.Parallel()

	p := &packet.GetIdentityRequest{}
	if got := p.Type(); got != packet.TypeGetIdentityRequest {
		t.Fatalf("Type() mismatch: \n\tgot: 0x%02x\n\twant: 0x%02x", got, packet.TypeGetIdentityRequest)
	}
}

func TestGetIdentityRequest_ExpectedLen(t *testing.T) {
	t.Parallel()

	p := &packet.GetIdentityRequest{}
	got, ok := p.ExpectedLen()

	if !ok {
		t.Fatalf("ExpectedLen() ok mismatch:\n\tgot:  %v\n\twant: true", ok)
	}
	if got != 0 {
		t.Fatalf("ExpectedLen() len mismatch:\n\tgot:  %d\n\twant: 0", got)
	}
}

func TestGetIdentityRequest_Encode_NoError(t *testing.T) {
	t.Parallel()

	p := &packet.GetIdentityRequest{}
	var buf bytes.Buffer

	if err := p.Encode(&buf); err != nil {
		t.Fatalf("Encode() unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("Encode() wrote data:\n\tgot:  %d bytes\n\twant: 0 bytes", buf.Len())
	}
}

func TestGetIdentityRequest_Decode_NoError(t *testing.T) {
	t.Parallel()

	p := &packet.GetIdentityRequest{}
	r := bytes.NewReader([]byte{0xAA, 0xBB, 0xCC})

	if err := p.Decode(r); err != nil {
		t.Fatalf("Decode() unexpected error: %v", err)
	}
}
