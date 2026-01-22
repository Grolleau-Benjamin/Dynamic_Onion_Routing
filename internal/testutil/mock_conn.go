package testutil

import (
	"bytes"
	"net"
	"sync"
	"time"
)

type MockConn struct {
	ReadBuf    *bytes.Buffer
	WriteBuf   *bytes.Buffer
	closed     bool
	closedMu   sync.Mutex
	remoteAddr net.Addr
	ReadErr    error
	WriteErr   error
	CloseErr   error
}

func NewMockConn(data []byte) *MockConn {
	return &MockConn{
		ReadBuf:    bytes.NewBuffer(data),
		WriteBuf:   &bytes.Buffer{},
		remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345},
	}
}

func NewMockConnWithAddr(data []byte, addr net.Addr) *MockConn {
	return &MockConn{
		ReadBuf:    bytes.NewBuffer(data),
		WriteBuf:   &bytes.Buffer{},
		remoteAddr: addr,
	}
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	if m.ReadErr != nil {
		return 0, m.ReadErr
	}
	return m.ReadBuf.Read(b)
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	if m.WriteErr != nil {
		return 0, m.WriteErr
	}
	return m.WriteBuf.Write(b)
}

func (m *MockConn) Close() error {
	m.closedMu.Lock()
	defer m.closedMu.Unlock()
	m.closed = true
	return m.CloseErr
}

func (m *MockConn) IsClosed() bool {
	m.closedMu.Lock()
	defer m.closedMu.Unlock()
	return m.closed
}

func (m *MockConn) LocalAddr() net.Addr {
	return m.remoteAddr
}

func (m *MockConn) RemoteAddr() net.Addr {
	return m.remoteAddr
}

func (m *MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}
