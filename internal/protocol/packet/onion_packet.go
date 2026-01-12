package packet

import (
	"io"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/onion"
)

type OnionPacket struct {
	Data [onion.PacketSize]byte
}

func (pkt *OnionPacket) Type() uint8 {
	return TypeOnionPacket
}

func (pkt *OnionPacket) Encode(w io.Writer) error {
	_, err := w.Write(pkt.Data[:])
	return err
}

func (pkt *OnionPacket) Decode(r io.Reader) error {
	_, err := io.ReadFull(r, pkt.Data[:])
	return err
}

func (pkt *OnionPacket) ExpectedLen() (int, bool) {
	return onion.PacketSize, true
}
