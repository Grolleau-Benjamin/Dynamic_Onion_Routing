package transport

import (
	"net"
	"testing"
	"time"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/packet"
)

func TestNewTransport(t *testing.T) {
	t.Parallel()

	tr := NewTransport()

	if tr.dialTimeout != 3*time.Second {
		t.Errorf("dialTimeout mismatch:\n\tgot:  %v\n\twant: %v",
			tr.dialTimeout, 3*time.Second)
	}

	if tr.writeTimeout != 2*time.Second {
		t.Errorf("writeTimeout mismatch:\n\tgot:  %v\n\twant: %v",
			tr.writeTimeout, 2*time.Second)
	}

	if tr.readTimeout != 5*time.Second {
		t.Errorf("readTimeout mismatch:\n\tgot:  %v\n\twant: %v",
			tr.readTimeout, 5*time.Second)
	}
}

func TestDialEndpoint_InvalidEndpoint(t *testing.T) {
	t.Parallel()

	ep := identity.Endpoint{
		IP:   net.ParseIP("192.0.2.1"),
		Port: 9999,
	}

	timeout := 100 * time.Millisecond

	_, err := dialEndpoint(ep, timeout)
	if err == nil {
		t.Error("expected error when dialing unreachable endpoint")
	}
}

func TestDialEndpoint_Timeout(t *testing.T) {
	t.Parallel()

	ep := identity.Endpoint{
		IP:   net.ParseIP("192.0.2.1"),
		Port: 9999,
	}

	timeout := 1 * time.Millisecond
	start := time.Now()

	_, err := dialEndpoint(ep, timeout)

	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected timeout error")
	}

	if elapsed > 2*time.Second {
		t.Errorf("dial took too long: %v (expected timeout around %v)", elapsed, timeout)
	}
}

func TestTransport_Send_InvalidEndpoint(t *testing.T) {
	t.Parallel()

	tr := NewTransport()

	ep := identity.Endpoint{
		IP:   net.ParseIP("192.0.2.1"),
		Port: 9999,
	}

	pkt := &packet.GetIdentityRequest{}

	err := tr.Send(ep, pkt)
	if err == nil {
		t.Error("expected error when sending to unreachable endpoint")
	}
}

func TestTransport_Request_InvalidEndpoint(t *testing.T) {
	t.Parallel()

	tr := NewTransport()

	ep := identity.Endpoint{
		IP:   net.ParseIP("192.0.2.1"),
		Port: 9999,
	}

	req := &packet.GetIdentityRequest{}

	_, err := tr.Request(ep, req)
	if err == nil {
		t.Error("expected error when requesting from unreachable endpoint")
	}
}

func TestTransport_Send_WithMockServer(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		_, _ = packet.ReadPacket(conn)
	}()

	addr := listener.Addr().(*net.TCPAddr)
	ep := identity.Endpoint{
		IP:   addr.IP,
		Port: uint16(addr.Port),
	}

	tr := NewTransport()
	pkt := &packet.GetIdentityRequest{}

	err = tr.Send(ep, pkt)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}

	<-done
}

func TestTransport_Request_WithMockServer(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	testUUID := [16]byte{0x01, 0x02, 0x03, 0x04}
	testPubKey := [32]byte{0xaa, 0xbb, 0xcc, 0xdd}

	done := make(chan error)
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			done <- err
			return
		}
		defer func() { _ = conn.Close() }()

		_, err = packet.ReadPacket(conn)
		if err != nil {
			done <- err
			return
		}

		resp := &packet.GetIdentityResponse{
			Ruuid:     testUUID,
			PublicKey: testPubKey,
		}

		err = packet.WritePacket(conn, resp)
		if err != nil {
			done <- err
			return
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	ep := identity.Endpoint{
		IP:   addr.IP,
		Port: uint16(addr.Port),
	}

	tr := NewTransport()
	req := &packet.GetIdentityRequest{}

	resp, err := tr.Request(ep, req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	identityResp, ok := resp.(*packet.GetIdentityResponse)
	if !ok {
		t.Fatalf("expected *GetIdentityResponse, got %T", resp)
	}

	if identityResp.Ruuid != testUUID {
		t.Errorf("UUID mismatch:\n\tgot:  %x\n\twant: %x",
			identityResp.Ruuid, testUUID)
	}

	if identityResp.PublicKey != testPubKey {
		t.Errorf("PubKey mismatch:\n\tgot:  %x\n\twant: %x",
			identityResp.PublicKey, testPubKey)
	}

	serverErr := <-done
	if serverErr != nil {
		t.Errorf("mock server error: %v", serverErr)
	}
}

func TestTransport_Request_ServerClosesEarly(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		_ = conn.Close()
	}()

	addr := listener.Addr().(*net.TCPAddr)
	ep := identity.Endpoint{
		IP:   addr.IP,
		Port: uint16(addr.Port),
	}

	tr := NewTransport()
	req := &packet.GetIdentityRequest{}

	_, err = tr.Request(ep, req)
	if err == nil {
		t.Error("expected error when server closes early")
	}
}

func TestTransport_Send_MultiplePackets(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	packetsReceived := 0
	done := make(chan struct{})

	go func() {
		defer close(done)
		for range 3 {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			_, err = packet.ReadPacket(conn)
			if err == nil {
				packetsReceived++
			}
			_ = conn.Close()
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	ep := identity.Endpoint{
		IP:   addr.IP,
		Port: uint16(addr.Port),
	}

	tr := NewTransport()
	pkt := &packet.GetIdentityRequest{}

	for i := range 3 {
		err := tr.Send(ep, pkt)
		if err != nil {
			t.Errorf("Send %d failed: %v", i, err)
		}
	}

	<-done

	if packetsReceived != 3 {
		t.Errorf("packets received mismatch:\n\tgot:  %d\n\twant: 3",
			packetsReceived)
	}
}

func TestTransport_Dial_ReusesConnection(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	addr := listener.Addr().(*net.TCPAddr)
	ep := identity.Endpoint{
		IP:   addr.IP,
		Port: uint16(addr.Port),
	}

	tr := NewTransport()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	conn1, err := tr.dial(ep)
	if err != nil {
		t.Fatalf("first dial failed: %v", err)
	}
	_ = conn1.Close()

	conn2, err := tr.dial(ep)
	if err != nil {
		t.Fatalf("second dial failed: %v", err)
	}
	_ = conn2.Close()
}

func BenchmarkTransport_Send(b *testing.B) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("failed to create listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer func() { _ = c.Close() }()
				_, _ = packet.ReadPacket(c)
			}(conn)
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	ep := identity.Endpoint{
		IP:   addr.IP,
		Port: uint16(addr.Port),
	}

	tr := NewTransport()
	pkt := &packet.GetIdentityRequest{}

	b.ResetTimer()
	for b.Loop() {
		_ = tr.Send(ep, pkt)
	}
}

func BenchmarkTransport_Request(b *testing.B) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("failed to create listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	testUUID := [16]byte{0x01, 0x02, 0x03, 0x04}
	testPubKey := [32]byte{0xaa, 0xbb, 0xcc, 0xdd}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer func() { _ = c.Close() }()
				_, _ = packet.ReadPacket(c)

				resp := &packet.GetIdentityResponse{
					Ruuid:     testUUID,
					PublicKey: testPubKey,
				}
				_ = packet.WritePacket(c, resp)
			}(conn)
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	ep := identity.Endpoint{
		IP:   addr.IP,
		Port: uint16(addr.Port),
	}

	tr := NewTransport()
	req := &packet.GetIdentityRequest{}

	b.ResetTimer()
	for b.Loop() {
		_, _ = tr.Request(ep, req)
	}
}
