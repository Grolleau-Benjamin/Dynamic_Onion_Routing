package testutil

import (
	"bytes"
	"net"
	"sync"
	"time"
)

type MockConn struct {
	mu         sync.Mutex
	ReadBuf    *bytes.Buffer
	WriteBuf   *bytes.Buffer
	closed     bool
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
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return 0, net.ErrClosed
	}
	if m.ReadErr != nil {
		return 0, m.ReadErr
	}
	return m.ReadBuf.Read(b)
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return 0, net.ErrClosed
	}
	if m.WriteErr != nil {
		return 0, m.WriteErr
	}
	return m.WriteBuf.Write(b)
}

func (m *MockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return m.CloseErr
}

func (m *MockConn) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

func (m *MockConn) GetWrittenBytes() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]byte, m.WriteBuf.Len())
	copy(out, m.WriteBuf.Bytes())
	return out
}

func (m *MockConn) GetWrittenLen() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.WriteBuf.Len()
}

func (m *MockConn) ResetWriteBuf() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WriteBuf.Reset()
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
