package packet

import (
	"reflect"
	"testing"
)

func TestRegistry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		t    uint8
		want Packet
	}{
		{
			name: "TypeGetIdentityRequest",
			t:    TypeGetIdentityRequest,
			want: &GetIdentityRequest{},
		},
		{
			name: "TypeGetIdentityResponse",
			t:    TypeGetIdentityResponse,
			want: &GetIdentityResponse{},
		},
		{
			name: "TypeOnion",
			t:    TypeOnionPacket,
			want: &OnionPacket{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, ok := registry[tt.t]
			if !ok {
				t.Fatalf("no factory for type %#x", tt.t)
			}

			pkt := factory()
			if reflect.TypeOf(pkt) != reflect.TypeOf(tt.want) {
				t.Fatalf("wrong packet type for %#x: \n\tgot %T \n\twant %T",
					tt.t, pkt, tt.want)
			}
		})
	}
}
