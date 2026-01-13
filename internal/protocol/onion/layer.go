// Package onion defines the Onion routing layer format used by DOR.
package onion

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/crypto"
)

const (
	MaxWrappedKey  = 3
	WrappedKeySize = 76

	// EPK (32) + WrappedKeys + Flags (1) + PayloadNonce (12) + CypherTextLen
	FixedHeaderSize = 32 + (MaxWrappedKey * WrappedKeySize) + 1 + 12 + 2
)

var (
	HKDFInfoCipherTextLenMask = []byte("DORv1:CipherTextLenMask16")
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
// |PN [11] | CT len XOR (2o) |        |
// +--------+--------+--------+        |
// ~          Cipher Text              ~
// |                                   |
// +--------+--------+--------+--------+
//
// EPK              -> Ephemeral Public Key (32 bytes)
// WrappedKeys      -> 3 slots of (Nonce[12] + Cipher[64])
// Flags            -> Layer Flags (1 byte)
// PayloadNonce     -> ChaCha20 Nonce for CipherText (12 bytes)
// CipherTextLenXor -> uint16(len(CipherText)) XOR mask16 (2 bytes, BE)
// CipherText       -> AEAD ciphertext (includes Poly1305 tag), followed by random padding up to PacketSize

type OnionLayer struct {
	EPK              [32]byte
	WrappedKeys      [MaxWrappedKey]WrappedKey
	Flags            uint8 // not used yet, reserved
	PayloadNonce     [12]byte
	CipherTextLenXor uint16
	CipherText       []byte
}

func CipherTextLenMask16(cipherKey [32]byte, payloadNonce [12]byte) (uint16, error) {
	maskBytes, err := crypto.HKDFSha256(
		cipherKey[:],
		payloadNonce[:],           // salt
		HKDFInfoCipherTextLenMask, // info
	)
	if err != nil {
		return 0, err
	}
	if len(maskBytes) < 2 {
		return 0, fmt.Errorf("hkdf output too short: %d", len(maskBytes))
	}
	return binary.BigEndian.Uint16(maskBytes[:2]), nil
}

func (ol *OnionLayer) HeaderBytes() ([]byte, error) {
	out := make([]byte, 0, FixedHeaderSize)

	out = append(out, ol.EPK[:]...)

	for _, wk := range ol.WrappedKeys {
		out = append(out, wk.Nonce[:]...)
		out = append(out, wk.CipherText[:]...)
	}

	out = append(out, ol.Flags)
	out = append(out, ol.PayloadNonce[:]...)

	var bufCipherLen [2]byte
	binary.BigEndian.PutUint16(bufCipherLen[:], ol.CipherTextLenXor)
	out = append(out, bufCipherLen[:]...)

	return out, nil
}

func (ol *OnionLayer) Bytes() ([]byte, error) {
	header, err := ol.HeaderBytes()
	if err != nil {
		return nil, err
	}

	out := make([]byte, 0, len(header)+len(ol.CipherText))
	out = append(out, header...)
	out = append(out, ol.CipherText...)

	return out, nil
}

func (ol *OnionLayer) BytesPadded() ([]byte, error) {
	rawBytes, err := ol.Bytes()
	if err != nil {
		return nil, err
	}

	currentSize := len(rawBytes)
	if currentSize > PacketSize {
		return nil, fmt.Errorf("packet overflow: layer size %d exceeds limit %d", currentSize, PacketSize)
	}

	out := make([]byte, PacketSize)
	copy(out, rawBytes)

	paddingStart := currentSize
	if paddingStart < PacketSize {
		if _, err := rand.Read(out[paddingStart:]); err != nil {
			return nil, fmt.Errorf("failed to generate random padding: %v", err)
		}
	}

	return out, nil
}

func (ol *OnionLayer) Parse(data []byte) error {
	if len(data) < FixedHeaderSize {
		return fmt.Errorf("data too short: need at least %d bytes for header, got %d",
			FixedHeaderSize, len(data))
	}

	offset := 0

	copy(ol.EPK[:], data[offset:offset+32])
	offset += 32

	for i := range MaxWrappedKey {
		copy(ol.WrappedKeys[i].Nonce[:], data[offset:offset+12])
		offset += 12

		copy(ol.WrappedKeys[i].CipherText[:], data[offset:offset+64])
		offset += 64
	}

	ol.Flags = data[offset]
	offset++

	copy(ol.PayloadNonce[:], data[offset:offset+12])
	offset += 12

	ol.CipherTextLenXor = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	ol.CipherText = make([]byte, len(data)-offset)
	copy(ol.CipherText, data[offset:])

	return nil
}

func (ol *OnionLayer) CipherTextLen(cipherKey [32]byte) (int, error) {
	mask, err := CipherTextLenMask16(cipherKey, ol.PayloadNonce)
	if err != nil {
		return 0, err
	}
	realLen := ol.CipherTextLenXor ^ mask
	return int(realLen), nil
}

func (ol *OnionLayer) TrimCipherText(cipherKey [32]byte) error {
	realLen, err := ol.CipherTextLen(cipherKey)
	if err != nil {
		return err
	}
	if realLen > len(ol.CipherText) {
		return fmt.Errorf("invalid cipher text length: %d exceeds available data %d",
			realLen, len(ol.CipherText))
	}
	ol.CipherText = ol.CipherText[:realLen]
	return nil
}
