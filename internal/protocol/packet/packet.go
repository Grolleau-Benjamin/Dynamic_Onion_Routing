package packet

import "io"

const (
	TypeGetIdentityRequest  uint8 = 0x00
	TypeGetIdentityResponse uint8 = 0x01
)

type Packet interface {
	Type() uint8

	Encode(w io.Writer) error
	Decode(r io.Reader) error

	ExpectedLen() (int, bool)
}
