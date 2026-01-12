package onion

import (
	"encoding/binary"
	"fmt"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
)

const (
	// RRRRlnnh (1) + PayloadLength (2)
	InnerMetadataFixedSize = 1 + 2
)

// OnionLayerCiphered is the logical decrypted content of an OnionLayer.CipherText
// It is NEVER transmitted as-is.
type OnionLayerCiphered struct {
	LastServer        bool
	NextHops          []identity.Endpoint
	UtilPayloadLength uint16 // Actual payload length before padding
	Payload           []byte
}

// 0        7        15       23       31
// +--------+--------+--------+--------+
// |RRRRlnnh|   Payload Len   | NNH[0] |
// +--------+--------+--------+--------+
// |                                   |
// ~ Next Hops List (Variable) [1:...] ~
// |                                   |
// +--------+--------+--------+--------+
// |                                   |
// ~          Actual Payload           ~
// |                                   |
// +--------+--------+--------+--------+
//
// R        -> Reserved (4 bits)
// l        -> LastServer (1 bit)
// nnh      -> Nb NextHops (3 bits)

func (olc *OnionLayerCiphered) Bytes() ([]byte, error) {
	headerLen := InnerMetadataFixedSize
	for _, ep := range olc.NextHops {
		headerLen += ep.BytesLen()
	}

	out := make([]byte, 0, headerLen+len(olc.Payload))
	var flags uint8
	if olc.LastServer {
		flags |= FlagLastServer
	}

	if len(olc.NextHops) >= MaxWrappedKey {
		return nil, fmt.Errorf("too mutch wrappedKeys, max is %d", MaxWrappedKey)
	}
	flags |= uint8(len(olc.NextHops)) & FlagNbNextHops
	out = append(out, flags)

	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], olc.UtilPayloadLength)
	out = append(out, buf[:]...)

	for _, ep := range olc.NextHops {
		epBytes, err := ep.Bytes()
		if err != nil {
			return nil, err
		}
		out = append(out, epBytes...)
	}

	out = append(out, olc.Payload...)

	return out, nil
}
