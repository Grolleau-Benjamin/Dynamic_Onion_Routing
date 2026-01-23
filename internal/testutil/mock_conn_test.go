package testutil

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestNewMockConn(t *testing.T) {
	t.Parallel()

	data := []byte{0x01, 0x02, 0x03, 0x04}
	conn := NewMockConn(data)

	if conn.ReadBuf == nil {
		t.Fatalf("ReadBuf should not be nil")
	}

	if conn.WriteBuf == nil {
		t.Fatalf("WriteBuf should not be nil")
	}

	if conn.remoteAddr == nil {
		t.Fatalf("remoteAddr should not be nil")
	}

	readData, err := io.ReadAll(conn.ReadBuf)
	if err != nil {
		t.Fatalf("failed to read from ReadBuf: %v", err)
	}

	if !bytes.Equal(readData, data) {
		t.Fatalf("ReadBuf data mismatch:\n\tgot:  %x\n\twant: %x", readData, data)
	}

	expectedAddr := "127.0.0.1:12345"
	if conn.remoteAddr.String() != expectedAddr {
		t.Fatalf("remoteAddr mismatch:\n\tgot:  %s\n\twant: %s",
			conn.remoteAddr.String(), expectedAddr)
	}
}

func TestNewMockConnWithAddr(t *testing.T) {
	t.Parallel()

	data := []byte{0xaa, 0xbb, 0xcc}
	customAddr := &net.TCPAddr{IP: net.ParseIP("192.168.1.100"), Port: 8080}

	conn := NewMockConnWithAddr(data, customAddr)

	if conn.RemoteAddr() != customAddr {
		t.Fatalf("remoteAddr mismatch:\n\tgot:  %v\n\twant: %v",
			conn.RemoteAddr(), customAddr)
	}

	readData, err := io.ReadAll(conn.ReadBuf)
	if err != nil {
		t.Fatalf("failed to read from ReadBuf: %v", err)
	}

	if !bytes.Equal(readData, data) {
		t.Fatalf("ReadBuf data mismatch:\n\tgot:  %x\n\twant: %x", readData, data)
	}
}

func TestMockConn_Read(t *testing.T) {
	t.Parallel()

	data := []byte{0x01, 0x02, 0x03, 0x04}
	conn := NewMockConn(data)

	buf := make([]byte, 4)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if n != 4 {
		t.Fatalf("bytes read mismatch:\n\tgot:  %d\n\twant: 4", n)
	}

	if !bytes.Equal(buf, data) {
		t.Fatalf("read data mismatch:\n\tgot:  %x\n\twant: %x", buf, data)
	}
}

func TestMockConn_Read_EOF(t *testing.T) {
	t.Parallel()

	conn := NewMockConn([]byte{})

	buf := make([]byte, 10)
	n, err := conn.Read(buf)

	if err != io.EOF {
		t.Fatalf("expected EOF, got: %v", err)
	}

	if n != 0 {
		t.Fatalf("expected 0 bytes read, got: %d", n)
	}
}

func TestMockConn_Read_With_Error(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("mock read error")
	conn := NewMockConn([]byte{0x01, 0x02})
	conn.ReadErr = expectedErr

	buf := make([]byte, 10)
	n, err := conn.Read(buf)

	if err != expectedErr {
		t.Fatalf("error mismatch:\n\tgot:  %v\n\twant: %v", err, expectedErr)
	}

	if n != 0 {
		t.Fatalf("expected 0 bytes read on error, got: %d", n)
	}
}

func TestMockConn_Write(t *testing.T) {
	t.Parallel()

	conn := NewMockConn([]byte{})
	data := []byte{0xaa, 0xbb, 0xcc, 0xdd}

	n, err := conn.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if n != 4 {
		t.Fatalf("bytes written mismatch:\n\tgot:  %d\n\twant: 4", n)
	}

	written := conn.WriteBuf.Bytes()
	if !bytes.Equal(written, data) {
		t.Fatalf("written data mismatch:\n\tgot:  %x\n\twant: %x", written, data)
	}
}

func TestMockConn_Write_With_Error(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("mock write error")
	conn := NewMockConn([]byte{})
	conn.WriteErr = expectedErr

	data := []byte{0x01, 0x02}
	n, err := conn.Write(data)

	if err != expectedErr {
		t.Fatalf("error mismatch:\n\tgot:  %v\n\twant: %v", err, expectedErr)
	}

	if n != 0 {
		t.Fatalf("expected 0 bytes written on error, got: %d", n)
	}

	if conn.WriteBuf.Len() != 0 {
		t.Fatalf("WriteBuf should be empty on error, got %d bytes", conn.WriteBuf.Len())
	}
}

func TestMockConn_Close(t *testing.T) {
	t.Parallel()

	conn := NewMockConn([]byte{})

	if conn.IsClosed() {
		t.Fatalf("connection should not be closed initially")
	}

	err := conn.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if !conn.IsClosed() {
		t.Fatalf("connection should be closed after Close()")
	}
}

func TestMockConn_Close_With_Error(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("mock close error")
	conn := NewMockConn([]byte{})
	conn.CloseErr = expectedErr

	err := conn.Close()
	if err != expectedErr {
		t.Fatalf("error mismatch:\n\tgot:  %v\n\twant: %v", err, expectedErr)
	}

	if !conn.IsClosed() {
		t.Fatalf("connection should be marked closed even with error")
	}
}

func TestMockConn_Close_Multiple(t *testing.T) {
	t.Parallel()

	conn := NewMockConn([]byte{})

	err := conn.Close()
	if err != nil {
		t.Fatalf("first Close failed: %v", err)
	}

	err = conn.Close()
	if err != nil {
		t.Fatalf("second Close failed: %v", err)
	}

	if !conn.IsClosed() {
		t.Fatalf("connection should remain closed")
	}
}

func TestMockConn_IsClosed_ThreadSafe(t *testing.T) {
	t.Parallel()

	conn := NewMockConn([]byte{})

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = conn.IsClosed()
		}()
	}

	wg.Wait()
}

func TestMockConn_Close_ThreadSafe(t *testing.T) {
	t.Parallel()

	conn := NewMockConn([]byte{})

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = conn.Close()
		}()
	}

	wg.Wait()

	if !conn.IsClosed() {
		t.Fatalf("connection should be closed after concurrent Close calls")
	}
}

func TestMockConn_LocalAddr(t *testing.T) {
	t.Parallel()

	expectedAddr := &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 9000}
	conn := NewMockConnWithAddr([]byte{}, expectedAddr)

	localAddr := conn.LocalAddr()
	if localAddr != expectedAddr {
		t.Fatalf("LocalAddr mismatch:\n\tgot:  %v\n\twant: %v",
			localAddr, expectedAddr)
	}
}

func TestMockConn_RemoteAddr(t *testing.T) {
	t.Parallel()

	expectedAddr := &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 8080}
	conn := NewMockConnWithAddr([]byte{}, expectedAddr)

	remoteAddr := conn.RemoteAddr()
	if remoteAddr != expectedAddr {
		t.Fatalf("RemoteAddr mismatch:\n\tgot:  %v\n\twant: %v",
			remoteAddr, expectedAddr)
	}
}

func TestMockConn_SetDeadline(t *testing.T) {
	t.Parallel()

	conn := NewMockConn([]byte{})
	deadline := time.Now().Add(5 * time.Second)

	err := conn.SetDeadline(deadline)
	if err != nil {
		t.Fatalf("SetDeadline failed: %v", err)
	}
}

func TestMockConn_SetReadDeadline(t *testing.T) {
	t.Parallel()

	conn := NewMockConn([]byte{})
	deadline := time.Now().Add(5 * time.Second)

	err := conn.SetReadDeadline(deadline)
	if err != nil {
		t.Fatalf("SetReadDeadline failed: %v", err)
	}
}

func TestMockConn_SetWriteDeadline(t *testing.T) {
	t.Parallel()

	conn := NewMockConn([]byte{})
	deadline := time.Now().Add(5 * time.Second)

	err := conn.SetWriteDeadline(deadline)
	if err != nil {
		t.Fatalf("SetWriteDeadline failed: %v", err)
	}
}

func TestMockConn_ReadWrite_Sequence(t *testing.T) {
	t.Parallel()

	initialData := []byte{0x01, 0x02, 0x03}
	conn := NewMockConn(initialData)

	readBuf := make([]byte, 3)
	n, err := conn.Read(readBuf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != 3 {
		t.Fatalf("Read bytes mismatch: got %d, want 3", n)
	}

	writeData := []byte{0xaa, 0xbb}
	n, err = conn.Write(writeData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 2 {
		t.Fatalf("Write bytes mismatch: got %d, want 2", n)
	}

	if !bytes.Equal(conn.WriteBuf.Bytes(), writeData) {
		t.Fatalf("WriteBuf mismatch:\n\tgot:  %x\n\twant: %x",
			conn.WriteBuf.Bytes(), writeData)
	}
}

func TestMockConn_ImplementsNetConn(t *testing.T) {
	t.Parallel()

	var _ net.Conn = (*MockConn)(nil)
}
