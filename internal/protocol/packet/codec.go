package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

func ReadPacket(r io.Reader) (Packet, error) {
	header := make([]byte, 3)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	packetType := header[0]
	length := int(binary.BigEndian.Uint16(header[1:3]))

	factory, ok := registry[packetType]
	if !ok {
		return nil, fmt.Errorf("unknown packet type: 0x%02x", packetType)
	}
	p := factory()

	if exp, ok := p.ExpectedLen(); ok && exp != length {
		return nil, fmt.Errorf(
			"invalid payload length for packet 0x%02x: got %d, expected %d",
			packetType, length, exp,
		)
	}

	lr := io.LimitReader(r, int64(length))
	if err := p.Decode(lr); err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	if n, err := io.Copy(io.Discard, lr); err != nil {
		return nil, err
	} else if n != 0 {
		return nil, fmt.Errorf("packet payload not fully consumed")
	}

	return p, nil
}

func WritePacket(w io.Writer, p Packet) error {
	var buf bytes.Buffer
	if err := p.Encode(&buf); err != nil {
		return fmt.Errorf("failed to encode packet payload: %w", err)
	}

	payloadLen := buf.Len()
	if payloadLen > 65535 {
		return fmt.Errorf("packet too large: %d bytes (max 65535)", payloadLen)
	}

	header := make([]byte, 3)
	header[0] = p.Type()
	binary.BigEndian.PutUint16(header[1:3], uint16(payloadLen))

	if _, err := w.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := buf.WriteTo(w); err != nil {
		return fmt.Errorf("failed to write payload: %w", err)
	}

	return nil
}
