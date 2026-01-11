package onion

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
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
// |                                   |
// ~           Random Padding          ~
// ~ (align to multiple of 1024 bytes) ~
// |                                   |
// +--------+--------+--------+--------+
//
// R        -> Reserved (4 bits)
// l        -> LastServer (1 bit)
// nnh      -> Nb NextHops (3 bits)
// Total Packet Size must be N * OnionBlockSize (N * 1024)

func (olc *OnionLayerCiphered) Bytes() ([]byte, error) {
	headerLen := 1 + 2
	for _, ep := range olc.NextHops {
		headerLen += ep.BytesLen()
	}

	out := make([]byte, 0, headerLen+len(olc.Payload))
	var flags uint8
	if olc.LastServer {
		flags |= FlagLastServer
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

	padLen := OnionBlockSize - (len(out) % OnionBlockSize)
	if padLen == OnionBlockSize {
		padLen = 0
	}

	if padLen > 0 {
		pad := make([]byte, padLen)
		if _, err := rand.Read(pad); err != nil {
			return nil, fmt.Errorf("random padding generation failed")
		}
		out = append(out, pad...)
	}

	return out, nil // TODO: add some tests for out%1024 == 0
}
