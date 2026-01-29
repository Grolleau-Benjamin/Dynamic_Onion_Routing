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
// |RRRRlnnh|   Payload Len   |  NH[0] |
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

	if len(olc.NextHops) > MaxWrappedKey {
		return nil, fmt.Errorf("too much wrappedKeys, max is %d", MaxWrappedKey)
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

func (olc *OnionLayerCiphered) Parse(data []byte) error {
	if len(data) < InnerMetadataFixedSize {
		return fmt.Errorf("buffer too short")
	}

	offset := 0
	flags := data[offset]
	olc.LastServer = (flags & FlagLastServer) != 0
	nnh := int(flags & FlagNbNextHops)
	offset++

	olc.UtilPayloadLength = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	olc.NextHops = make([]identity.Endpoint, nnh)
	for i := range nnh {
		if offset >= len(data) {
			return fmt.Errorf("unexpected EOF reading hop %d", i)
		}

		n, err := olc.NextHops[i].Parse(data[offset:])
		if err != nil {
			return fmt.Errorf("failed to parse hop %d: %w", i, err)
		}
		offset += n
	}

	if offset > len(data) {
		return fmt.Errorf("offset exceeds data length")
	}

	olc.Payload = make([]byte, len(data)-offset)
	copy(olc.Payload, data[offset:])

	return nil
}
