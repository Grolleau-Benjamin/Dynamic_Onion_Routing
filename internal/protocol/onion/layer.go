// Package onion defines the Onion routing layer format used by DOR.
package onion

const (
	MaxWrappedKey  = 3
	WrappedKeySize = 76

	OnionBlockSize = 1024

	// EPK (32) + WrappedKeys + Flags (1) + PayloadNonce (12)
	FixedHeaderSize = 32 + (MaxWrappedKey * WrappedKeySize) + 1 + 12
)

// 0        7        15       23       31
// +--------+--------+--------+--------+
// |                                   |
// ~           EPK (32 bytes)          ~
// |                                   |
// +--------+--------+--------+--------+
// |                                   |
// ~      WrappedKeys (Fixed List)     ~
// ~      MaxWrappedKey * 76 bytes     ~
// ~        (3 * 76 = 228 bytes)       ~
// |                                   |
// +--------+--------+--------+--------+
// |  Flags |   PayloadNonce [0..2]    |
// +--------+--------+--------+--------+
// |          PayloadNonce [3..6]      |
// +--------+--------+--------+--------+
// |          PayloadNonce [7..10]     |
// +--------+--------+--------+--------+
// |PN [11] |                          |
// +--------+                          |
// ~          Cipher Text              ~
// ~     (Multiple of OnionBlockSize)  ~
// |                                   |
// +--------+--------+--------+--------+
//
// EPK          -> Ephemeral Public Key (32 bytes)
// WrappedKeys  -> 3 slots of (Nonce[12] + Cipher[64])
// Flags        -> Layer Flags (1 byte)
// PayloadNonce -> ChaCha20 Nonce for CipherText (12 bytes)

type OnionLayer struct {
	EPK          [32]byte
	WrappedKeys  [MaxWrappedKey]WrappedKey
	Flags        uint8 // not used yet, reserved
	PayloadNonce [12]byte
	CipherText   []byte // Must be n * OnionBlockSize with n >= 1
}

func (ol *OnionLayer) Bytes() ([]byte, error) {
	out := make([]byte, 0, FixedHeaderSize+len(ol.CipherText))

	out = append(out, ol.EPK[:]...)

	for _, wk := range ol.WrappedKeys {
		out = append(out, wk.Nonce[:]...)
		out = append(out, wk.CipherText[:]...)
	}

	out = append(out, ol.Flags)
	out = append(out, ol.PayloadNonce[:]...)

	out = append(out, ol.CipherText...)

	return out, nil
}
