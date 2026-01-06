package packet

import "io"

type GetIdentityRequest struct{}

func (pkt *GetIdentityRequest) Type() uint8 {
	return TypeGetIdentityRequest
}

func (pkt *GetIdentityRequest) Encode(w io.Writer) error {
	return nil
}

func (pkt *GetIdentityRequest) Decode(r io.Reader) error {
	return nil
}

func (pkt *GetIdentityRequest) ExpectedLen() (int, bool) {
	return 0, true
}
