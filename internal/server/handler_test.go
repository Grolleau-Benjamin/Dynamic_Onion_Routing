package server

import (
	"errors"
	"io"
	"testing"

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

func TestServer_handleConn_DoubleClose(t *testing.T) {
	t.Parallel()

	conn := testutil.NewMockConn([]byte{})
	conn.ReadErr = io.EOF

	s := &Server{}
	s.handleConn(conn)

	if !conn.IsClosed() {
		t.Error("connection should be closed")
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
		if handler == nil {
			t.Errorf("handler for type 0x%02x is nil", typ)
		}
	}
}
