package packet

import "io"

type GetIdentityResponse struct {
	Ruuid     [16]byte
	PublicKey [32]byte
}

func (pkt *GetIdentityResponse) Type() uint8 {
	return TypeGetIdentityResponse
}

func (pkt *GetIdentityResponse) Encode(w io.Writer) error {
	if _, err := w.Write(pkt.Ruuid[:]); err != nil {
		return err
	}
	if _, err := w.Write(pkt.PublicKey[:]); err != nil {
		return err
	}
	return nil
}

func (pkt *GetIdentityResponse) Decode(r io.Reader) error {
	if _, err := io.ReadFull(r, pkt.Ruuid[:]); err != nil {
		return err
	}
	if _, err := io.ReadFull(r, pkt.PublicKey[:]); err != nil {
		return err
	}
	return nil
}

func (pkt *GetIdentityResponse) ExpectedLen() (int, bool) {
	return 48, true
}
