package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/testutil"
)

func TestServer_handleConn_EOF(t *testing.T) {
	t.Parallel()

	conn := testutil.NewMockConn([]byte{})
	conn.ReadErr = io.EOF

	s := &Server{}
	s.handleConn(conn)

	if !conn.IsClosed() {
		t.Error("connection should be closed after EOF")
	}
}

func TestServer_handleConn_ReadError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		readErr error
	}{
		{
			name:    "generic read error",
			readErr: errors.New("mock read error"),
		},
		{
			name:    "unexpected EOF",
			readErr: io.ErrUnexpectedEOF,
		},
		{
			name:    "closed pipe",
			readErr: io.ErrClosedPipe,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conn := testutil.NewMockConn([]byte{})
			conn.ReadErr = tt.readErr

			s := &Server{}
			s.handleConn(conn)

			if !conn.IsClosed() {
				t.Errorf("connection should be closed after error: %v", tt.readErr)
			}
		})
	}
}

func TestServer_handleConn_CloseError(t *testing.T) {
	t.Parallel()

	conn := testutil.NewMockConn([]byte{})
	conn.ReadErr = io.EOF
	conn.CloseErr = errors.New("mock close error")

	s := &Server{}
	s.handleConn(conn)

	if !conn.IsClosed() {
		t.Error("Close() should have been attempted")
	}
}

func TestHandlerRegistry_AllTypesRegistered(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		packetType uint8
		wantExists bool
	}{
		{
			name:       "GetIdentityRequest is registered",
			packetType: packet.TypeGetIdentityRequest,
			wantExists: true,
		},
		{
			name:       "OnionPacket is registered",
			packetType: packet.TypeOnionPacket,
			wantExists: true,
		},
		{
			name:       "unknown type is not registered",
			packetType: 0xFF,
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, exists := handlerRegistry[tt.packetType]
			if exists != tt.wantExists {
				t.Errorf("handlerRegistry[0x%02x] exists = %v, want %v",
					tt.packetType, exists, tt.wantExists)
			}
		})
	}
}

func TestHandlerRegistry_HandlersNotNil(t *testing.T) {
	t.Parallel()

	for typ, handler := range handlerRegistry {
		t.Run("", func(t *testing.T) {
			if handler == nil {
				t.Errorf("handler for type 0x%02x is nil", typ)
			}
		})
	}
}

func TestServer_handleConn_UnknownPacket(t *testing.T) {
	t.Parallel()

	client, server := net.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})

	s := &Server{}
	done := make(chan struct{})

	go func() {
		s.handleConn(server)
		close(done)
	}()

	_, err := client.Write([]byte{0xFE, 0x00, 0x00})
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_ = client.Close()

	<-done
}

func TestServer_handleConn_GetIdentityRequest(t *testing.T) {
	t.Parallel()

	testUUID := [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	testPubKey := [32]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22,
		0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00, 0x25, 0x06, 0x20, 0x03}

	s := &Server{
		Pi: &identity.PrivateIdentity{
			UUID:   testUUID,
			PubKey: testPubKey,
		},
	}

	packetData := []byte{packet.TypeGetIdentityRequest, 0x00, 0x00}
	conn := testutil.NewMockConn(packetData)

	done := make(chan struct{})
	go func() {
		s.handleConn(conn)
		close(done)
	}()

	<-done
	time.Sleep(50 * time.Millisecond)

	if conn.WriteBuf.Len() == 0 {
		t.Error("expected GetIdentityResponse to be written")
	}

	responseData := conn.WriteBuf.Bytes()

	if len(responseData) < 3 {
		t.Fatalf("response too short: %d bytes", len(responseData))
	}

	if responseData[0] != packet.TypeGetIdentityResponse {
		t.Errorf("wrong response type:\n\tgot:  0x%02x\n\twant: 0x%02x",
			responseData[0], packet.TypeGetIdentityResponse)
	}

	payloadLen := binary.BigEndian.Uint16(responseData[1:3])
	if payloadLen != 48 {
		t.Errorf("wrong payload length:\n\tgot:  %d\n\twant: 48", payloadLen)
	}

	body := responseData[3:]
	if !bytes.Equal(body[:16], testUUID[:]) {
		t.Errorf("UUID mismatch:\n\tgot:  %x\n\twant: %x", body[:16], testUUID[:])
	}

	if !bytes.Equal(body[16:48], testPubKey[:]) {
		t.Errorf("PubKey mismatch:\n\tgot:  %x\n\twant: %x", body[16:48], testPubKey[:])
	}

	if !conn.IsClosed() {
		t.Error("connection should be closed")
	}
}

func TestServer_handleConn_OnionPacket(t *testing.T) {
	t.Parallel()

	s := &Server{
		Pi: &identity.PrivateIdentity{
			UUID:   [16]byte{0x01, 0x02, 0x03, 0x04},
			PubKey: [32]byte{0xaa, 0xbb, 0xcc, 0xdd},
		},
	}

	payload := []byte{0x01, 0x02, 0x03, 0x04}
	packetData := make([]byte, 3+len(payload))
	packetData[0] = packet.TypeOnionPacket
	binary.BigEndian.PutUint16(packetData[1:3], uint16(len(payload)))
	copy(packetData[3:], payload)

	conn := testutil.NewMockConn(packetData)

	done := make(chan struct{})
	go func() {
		s.handleConn(conn)
		close(done)
	}()

	<-done

	if !conn.IsClosed() {
		t.Error("connection should be closed")
	}
}

func TestServer_handleConn_HandlerExecution(t *testing.T) {
	t.Parallel()

	validPacketData := []byte{packet.TypeGetIdentityRequest, 0x00, 0x00}
	conn := testutil.NewMockConn(validPacketData)

	s := &Server{}
	done := make(chan struct{})

	originalHandler := handlerRegistry[packet.TypeGetIdentityRequest]
	handlerRegistry[packet.TypeGetIdentityRequest] = func(p packet.Packet, c net.Conn, s *Server) {
		close(done)
	}
	defer func() { handlerRegistry[packet.TypeGetIdentityRequest] = originalHandler }()

	s.handleConn(conn)

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("handler was not executed within timeout")
	}

	if !conn.IsClosed() {
		t.Error("connection should be closed after loop finishes")
	}
}

func TestServer_handleConn_PanicRecovery(t *testing.T) {
	packetData := []byte{packet.TypeGetIdentityRequest, 0x00, 0x00}
	conn := testutil.NewMockConn(packetData)
	s := &Server{}

	handlerCalled := make(chan struct{})

	originalHandler := handlerRegistry[packet.TypeGetIdentityRequest]
	handlerRegistry[packet.TypeGetIdentityRequest] = func(p packet.Packet, c net.Conn, s *Server) {
		close(handlerCalled)
		panic("test panic")
	}
	defer func() { handlerRegistry[packet.TypeGetIdentityRequest] = originalHandler }()

	s.handleConn(conn)

	select {
	case <-handlerCalled:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("handler was not executed")
	}

	if !conn.IsClosed() {
		t.Error("connection should be closed")
	}
}

func TestServer_handleConn_MultiplePackets(t *testing.T) {
	t.Parallel()

	testUUID := [16]byte{0x01, 0x02, 0x03, 0x04}
	testPubKey := [32]byte{0xaa, 0xbb, 0xcc, 0xdd}

	s := &Server{
		Pi: &identity.PrivateIdentity{
			UUID:   testUUID,
			PubKey: testPubKey,
		},
	}

	numPackets := 3
	for i := range numPackets {
		packetData := []byte{packet.TypeGetIdentityRequest, 0x00, 0x00}
		conn := testutil.NewMockConn(packetData)

		done := make(chan struct{})
		go func() {
			s.handleConn(conn)
			close(done)
		}()

		<-done
		time.Sleep(50 * time.Millisecond)

		if !conn.IsClosed() {
			t.Errorf("connection %d should be closed", i)
		}

		if conn.WriteBuf.Len() == 0 {
			t.Errorf("connection %d: expected response to be written", i)
		}
	}
}

func TestServer_handleConn_ConcurrentHandlers(t *testing.T) {
	t.Parallel()

	testUUID := [16]byte{0x01, 0x02, 0x03, 0x04}
	testPubKey := [32]byte{0xaa, 0xbb, 0xcc, 0xdd}

	s := &Server{
		Pi: &identity.PrivateIdentity{
			UUID:   testUUID,
			PubKey: testPubKey,
		},
	}

	var wg sync.WaitGroup
	numRequests := 5

	for i := range numRequests {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			packetData := []byte{packet.TypeGetIdentityRequest, 0x00, 0x00}
			conn := testutil.NewMockConn(packetData)

			s.handleConn(conn)

			if !conn.IsClosed() {
				t.Errorf("connection %d should be closed", iteration)
			}
		}(i)
	}

	wg.Wait()
}
