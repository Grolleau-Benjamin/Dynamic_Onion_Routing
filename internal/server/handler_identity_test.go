package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/testutil"
)

func TestHandleGetIdentity_Success(t *testing.T) {
	t.Parallel()

	testUUID := [16]byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}
	testPubKey := [32]byte{
		0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22,
		0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00,
		0x25, 0x06, 0x20, 0x03, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	s := &Server{
		Pi: &identity.PrivateIdentity{
			UUID:   testUUID,
			PubKey: testPubKey,
		},
	}

	conn := testutil.NewMockConn([]byte{})
	pkt := &packet.GetIdentityRequest{}

	handleGetIdentity(pkt, conn, s)

	expectedResponsePacket := packet.GetIdentityResponse{}
	expectedPayloadSize, _ := expectedResponsePacket.ExpectedLen()
	expectedTotalLen := 3 + expectedPayloadSize

	if conn.WriteBuf.Len() != expectedTotalLen {
		t.Fatalf("response length mismatch:\n\tgot:  %d\n\twant: %d",
			conn.WriteBuf.Len(), expectedTotalLen)
	}

	responseData := conn.WriteBuf.Bytes()

	if responseData[0] != packet.TypeGetIdentityResponse {
		t.Errorf("wrong packet type:\n\tgot:  %x\n\twant: %x",
			responseData[0], packet.TypeGetIdentityResponse)
	}

	payloadLen := binary.BigEndian.Uint16(responseData[1:3])
	if int(payloadLen) != expectedPayloadSize {
		t.Errorf("wrong payload length in header:\n\tgot:  %d\n\twant: %d",
			payloadLen, expectedPayloadSize)
	}

	body := responseData[3:]
	if !bytes.Equal(body[:16], testUUID[:]) {
		t.Errorf("UUID mismatch:\n\tgot:  %x\n\twant: %x",
			body[:16], testUUID[:])
	}

	if !bytes.Equal(body[16:48], testPubKey[:]) {
		t.Errorf("PubKey mismatch:\n\tgot:  %x\n\twant: %x",
			body[16:48], testPubKey[:])
	}
}

func TestHandleGetIdentity_WriteError(t *testing.T) {
	t.Parallel()

	s := &Server{
		Pi: &identity.PrivateIdentity{
			UUID:   [16]byte{0x01, 0x02, 0x03, 0x04},
			PubKey: [32]byte{0xaa, 0xbb, 0xcc, 0xdd},
		},
	}

	conn := testutil.NewMockConn([]byte{})
	conn.WriteErr = errors.New("mock write error")

	pkt := &packet.GetIdentityRequest{}

	handleGetIdentity(pkt, conn, s)
}

func TestHandleGetIdentity_NetworkError(t *testing.T) {
	t.Parallel()

	s := &Server{
		Pi: &identity.PrivateIdentity{
			UUID:   [16]byte{0x01, 0x02, 0x03, 0x04},
			PubKey: [32]byte{0xaa, 0xbb, 0xcc, 0xdd},
		},
	}

	conn := testutil.NewMockConn([]byte{})
	conn.WriteErr = &net.OpError{
		Op:  "write",
		Net: "tcp",
		Err: errors.New("connection reset by peer"),
	}

	pkt := &packet.GetIdentityRequest{}

	handleGetIdentity(pkt, conn, s)
}

func TestHandleGetIdentity_ZeroValues(t *testing.T) {
	t.Parallel()

	testUUID := [16]byte{}
	testPubKey := [32]byte{}

	s := &Server{
		Pi: &identity.PrivateIdentity{
			UUID:   testUUID,
			PubKey: testPubKey,
		},
	}

	conn := testutil.NewMockConn([]byte{})
	pkt := &packet.GetIdentityRequest{}

	handleGetIdentity(pkt, conn, s)

	expectedResponsePacket := packet.GetIdentityResponse{}
	expectedPayloadSize, _ := expectedResponsePacket.ExpectedLen()
	expectedTotalLen := 3 + expectedPayloadSize

	if conn.WriteBuf.Len() != expectedTotalLen {
		t.Fatalf("response length mismatch:\n\tgot:  %d\n\twant: %d",
			conn.WriteBuf.Len(), expectedTotalLen)
	}

	responseData := conn.WriteBuf.Bytes()
	body := responseData[3:]

	if !bytes.Equal(body[:16], testUUID[:]) {
		t.Errorf("UUID mismatch:\n\tgot:  %x\n\twant: %x",
			body[:16], testUUID[:])
	}

	if !bytes.Equal(body[16:48], testPubKey[:]) {
		t.Errorf("PubKey mismatch:\n\tgot:  %x\n\twant: %x",
			body[16:48], testPubKey[:])
	}
}

func TestHandleGetIdentity_MultipleRequests(t *testing.T) {
	t.Parallel()

	testUUID := [16]byte{0x01, 0x02, 0x03, 0x04}
	testPubKey := [32]byte{0xaa, 0xbb, 0xcc, 0xdd}

	s := &Server{
		Pi: &identity.PrivateIdentity{
			UUID:   testUUID,
			PubKey: testPubKey,
		},
	}

	pkt := &packet.GetIdentityRequest{}
	expectedResponsePacket := packet.GetIdentityResponse{}
	expectedPayloadSize, _ := expectedResponsePacket.ExpectedLen()
	expectedTotalLen := 3 + expectedPayloadSize

	for i := range 5 {
		conn := testutil.NewMockConn([]byte{})
		handleGetIdentity(pkt, conn, s)

		if conn.WriteBuf.Len() != expectedTotalLen {
			t.Errorf("request %d: response length mismatch:\n\tgot:  %d\n\twant: %d",
				i, conn.WriteBuf.Len(), expectedTotalLen)
		}

		responseData := conn.WriteBuf.Bytes()
		body := responseData[3:]

		if !bytes.Equal(body[:16], testUUID[:]) {
			t.Errorf("request %d: UUID mismatch", i)
		}

		if !bytes.Equal(body[16:48], testPubKey[:]) {
			t.Errorf("request %d: PubKey mismatch", i)
		}
	}
}

func TestHandleGetIdentity_ConcurrentRequests(t *testing.T) {
	t.Parallel()

	testUUID := [16]byte{0x01, 0x02, 0x03, 0x04}
	testPubKey := [32]byte{0xaa, 0xbb, 0xcc, 0xdd}

	s := &Server{
		Pi: &identity.PrivateIdentity{
			UUID:   testUUID,
			PubKey: testPubKey,
		},
	}

	pkt := &packet.GetIdentityRequest{}
	expectedResponsePacket := packet.GetIdentityResponse{}
	expectedPayloadSize, _ := expectedResponsePacket.ExpectedLen()
	expectedTotalLen := 3 + expectedPayloadSize

	var wg sync.WaitGroup
	numRequests := 20

	for i := range numRequests {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			conn := testutil.NewMockConn([]byte{})
			handleGetIdentity(pkt, conn, s)

			if conn.WriteBuf.Len() != expectedTotalLen {
				t.Errorf("concurrent request %d: response length mismatch:\n\tgot:  %d\n\twant: %d",
					iteration, conn.WriteBuf.Len(), expectedTotalLen)
			}

			responseData := conn.WriteBuf.Bytes()
			body := responseData[3:]

			if !bytes.Equal(body[:16], testUUID[:]) {
				t.Errorf("concurrent request %d: UUID mismatch", iteration)
			}

			if !bytes.Equal(body[16:48], testPubKey[:]) {
				t.Errorf("concurrent request %d: PubKey mismatch", iteration)
			}
		}(i)
	}

	wg.Wait()
}

func TestHandleGetIdentity_IdentityNotModified(t *testing.T) {
	t.Parallel()

	originalUUID := [16]byte{0x01, 0x02, 0x03, 0x04}
	originalPubKey := [32]byte{0xaa, 0xbb, 0xcc, 0xdd}

	s := &Server{
		Pi: &identity.PrivateIdentity{
			UUID:   originalUUID,
			PubKey: originalPubKey,
		},
	}

	conn := testutil.NewMockConn([]byte{})
	pkt := &packet.GetIdentityRequest{}

	handleGetIdentity(pkt, conn, s)

	if s.Pi.UUID != originalUUID {
		t.Error("UUID was modified during request handling")
	}

	if s.Pi.PubKey != originalPubKey {
		t.Error("PubKey was modified during request handling")
	}
}

func TestHandleGetIdentity_DifferentRemoteAddresses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		addr net.Addr
	}{
		{
			name: "IPv4 address",
			addr: &net.TCPAddr{IP: net.ParseIP("192.168.1.100"), Port: 54321},
		},
		{
			name: "IPv6 address",
			addr: &net.TCPAddr{IP: net.ParseIP("::1"), Port: 8080},
		},
		{
			name: "localhost",
			addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345},
		},
	}

	testUUID := [16]byte{0x01, 0x02, 0x03, 0x04}
	testPubKey := [32]byte{0xaa, 0xbb, 0xcc, 0xdd}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &Server{
				Pi: &identity.PrivateIdentity{
					UUID:   testUUID,
					PubKey: testPubKey,
				},
			}

			conn := testutil.NewMockConnWithAddr([]byte{}, tt.addr)
			pkt := &packet.GetIdentityRequest{}

			handleGetIdentity(pkt, conn, s)

			expectedResponsePacket := packet.GetIdentityResponse{}
			expectedPayloadSize, _ := expectedResponsePacket.ExpectedLen()
			expectedTotalLen := 3 + expectedPayloadSize

			if conn.WriteBuf.Len() != expectedTotalLen {
				t.Errorf("response length mismatch for %s:\n\tgot:  %d\n\twant: %d",
					tt.addr, conn.WriteBuf.Len(), expectedTotalLen)
			}
		})
	}
}
