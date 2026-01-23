package server

import (
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/testutil"
)

func BenchmarkHandleGetIdentity(b *testing.B) {
	s := &Server{
		Pi: &identity.PrivateIdentity{
			UUID:   [16]byte{0x01, 0x02, 0x03, 0x04},
			PubKey: [32]byte{0xaa, 0xbb, 0xcc, 0xdd},
		},
	}

	pkt := &packet.GetIdentityRequest{}

	conn := testutil.NewMockConn([]byte{})

	b.ResetTimer()
	for b.Loop() {
		conn.ResetWriteBuf()
		handleGetIdentity(pkt, conn, s)
	}
}

func BenchmarkHandleGetIdentity_Parallel(b *testing.B) {
	s := &Server{
		Pi: &identity.PrivateIdentity{
			UUID:   [16]byte{0x01, 0x02, 0x03, 0x04},
			PubKey: [32]byte{0xaa, 0xbb, 0xcc, 0xdd},
		},
	}

	pkt := &packet.GetIdentityRequest{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		localConn := testutil.NewMockConn([]byte{})

		for pb.Next() {
			localConn.ResetWriteBuf()

			handleGetIdentity(pkt, localConn, s)
		}
	})
}
