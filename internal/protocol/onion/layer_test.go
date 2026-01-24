package onion_test

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/onion"
)

func TestCipherTextLenMask16(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		cipherKey    [32]byte
		payloadNonce [12]byte
		expected     uint16
		wantErr      bool
	}{
		{
			name: "valid input produces deterministic mask",
			cipherKey: [32]byte{
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
				0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
				0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
				0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			},
			payloadNonce: [12]byte{
				0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11,
				0x22, 0x33, 0x44, 0x55,
			},
			expected: 0xf107,
			wantErr:  false,
		},
		{
			name:         "zero values produce valid mask",
			cipherKey:    [32]byte{},
			payloadNonce: [12]byte{},
			expected:     0x98d1,
			wantErr:      false,
		},
		{
			name: "different nonce produces different mask",
			cipherKey: [32]byte{
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
				0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
				0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
				0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			},
			payloadNonce: [12]byte{
				0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
				0x99, 0xaa, 0xbb, 0xcc,
			},
			expected: 0x9f52,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := onion.CipherTextLenMask16(tt.cipherKey, tt.payloadNonce)

			if (err != nil) != tt.wantErr {
				t.Fatalf("CipherTextLenMask16() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.expected {
				t.Fatalf("CipherTextLenMask16() mismatch:\n\tgot:  0x%04x (%d)\n\twant: 0x%04x (%d)",
					got, got, tt.expected, tt.expected)
			}
		})
	}
}

func TestCipherTextLenMask16_Deterministic(t *testing.T) {
	t.Parallel()

	cipherKey := [32]byte{0x01, 0x02, 0x03, 0x04}
	payloadNonce := [12]byte{0xaa, 0xbb, 0xcc, 0xdd}

	mask1, err := onion.CipherTextLenMask16(cipherKey, payloadNonce)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	mask2, err := onion.CipherTextLenMask16(cipherKey, payloadNonce)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if mask1 != mask2 {
		t.Fatalf("function is not deterministic:\n\tfirst:  %d\n\tsecond: %d",
			mask1, mask2)
	}
}

func TestCipherTextLenMask16_DifferentInputs(t *testing.T) {
	t.Parallel()

	cipherKey := [32]byte{0x01, 0x02, 0x03, 0x04}
	nonce1 := [12]byte{0xaa, 0xbb, 0xcc, 0xdd}
	nonce2 := [12]byte{0x11, 0x22, 0x33, 0x44}

	mask1, err := onion.CipherTextLenMask16(cipherKey, nonce1)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	mask2, err := onion.CipherTextLenMask16(cipherKey, nonce2)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if mask1 == mask2 {
		t.Fatalf("different nonces should produce different masks")
	}
}

func TestOnionLayer_HeaderBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ol       onion.OnionLayer
		expected []byte
	}{
		{
			name: "valid input data",
			ol: onion.OnionLayer{
				EPK: [32]byte{
					0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
					0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
					0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
					0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
				},
				WrappedKeys: [3]onion.WrappedKey{},
				Flags:       0x00,
				PayloadNonce: [12]byte{
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11,
					0x22, 0x33, 0x44, 0x55,
				},
				CipherTextLenXor: 0xf107,
				CipherText:       []byte("A payload"),
			},
			expected: []byte{
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0xf1, 0x07,
			},
		},
		{
			name: "zero value data",
			ol: onion.OnionLayer{
				EPK:              [32]byte{},
				WrappedKeys:      [3]onion.WrappedKey{},
				Flags:            0x00,
				PayloadNonce:     [12]byte{},
				CipherTextLenXor: 0x0000,
				CipherText:       []byte{},
			},
			expected: make([]byte, onion.FixedHeaderSize),
		},
		{
			name: "max value data",
			ol: onion.OnionLayer{
				EPK: [32]byte{
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				},
				WrappedKeys: [3]onion.WrappedKey{
					{
						Nonce: [12]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						CipherText: [64]byte{
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff,
						},
					},
					{
						Nonce: [12]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						CipherText: [64]byte{
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff,
						},
					},
					{
						Nonce: [12]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						CipherText: [64]byte{
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
							0xff, 0xff, 0xff, 0xff,
						},
					},
				},
				Flags:            0xff,
				PayloadNonce:     [12]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				CipherTextLenXor: 0xffff,
				CipherText:       []byte{},
			},
			expected: func() []byte {
				buf := make([]byte, onion.FixedHeaderSize)
				for i := range buf {
					buf[i] = 0xff
				}
				return buf
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, _ := tt.ol.HeaderBytes()
			if !bytes.Equal(tt.expected, got) {
				t.Fatalf("HeaderBytes() invalid ouput.\n\tgot: %X\n\twant: %X", got, tt.expected)
			}
		})
	}
}

func TestOnionLayer_HeaderBytes_DifferentFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		flags uint8
	}{
		{name: "flag 0x00", flags: 0x00},
		{name: "flag 0x01", flags: 0x01},
		{name: "flag 0x80", flags: 0x80},
		{name: "flag 0xff", flags: 0xff},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ol := onion.OnionLayer{
				EPK:              [32]byte{0x01, 0x02, 0x03},
				WrappedKeys:      [3]onion.WrappedKey{},
				Flags:            tt.flags,
				PayloadNonce:     [12]byte{0xaa, 0xbb},
				CipherTextLenXor: 0x1234,
				CipherText:       []byte{},
			}

			got, err := ol.HeaderBytes()
			if err != nil {
				t.Fatalf("HeaderBytes() unexpected error: %v", err)
			}

			flagOffset := 32 + (onion.MaxWrappedKey * onion.WrappedKeySize)
			if got[flagOffset] != tt.flags {
				t.Fatalf("Flags mismatch at offset %d:\n\tgot:  0x%02x\n\twant: 0x%02x",
					flagOffset, got[flagOffset], tt.flags)
			}
		})
	}
}

func TestOnionLayer_HeaderBytes_DifferentCipherTextLenXor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		cipherTextLenXor uint16
	}{
		{name: "0x0000", cipherTextLenXor: 0x0000},
		{name: "0x0001", cipherTextLenXor: 0x0001},
		{name: "0x1234", cipherTextLenXor: 0x1234},
		{name: "0xffff", cipherTextLenXor: 0xffff},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ol := onion.OnionLayer{
				EPK:              [32]byte{},
				WrappedKeys:      [3]onion.WrappedKey{},
				Flags:            0x00,
				PayloadNonce:     [12]byte{},
				CipherTextLenXor: tt.cipherTextLenXor,
				CipherText:       []byte{},
			}

			got, err := ol.HeaderBytes()
			if err != nil {
				t.Fatalf("HeaderBytes() unexpected error: %v", err)
			}

			lenOffset := onion.FixedHeaderSize - 2
			gotLen := binary.BigEndian.Uint16(got[lenOffset:])

			if gotLen != tt.cipherTextLenXor {
				t.Fatalf("CipherTextLenXor mismatch:\n\tgot:  0x%04x\n\twant: 0x%04x",
					gotLen, tt.cipherTextLenXor)
			}
		})
	}
}

func TestOnionLayer_HeaderBytes_Idempotent(t *testing.T) {
	t.Parallel()

	ol := onion.OnionLayer{
		EPK:              [32]byte{0x01, 0x02, 0x03},
		WrappedKeys:      [3]onion.WrappedKey{},
		Flags:            0x00,
		PayloadNonce:     [12]byte{0xaa, 0xbb, 0xcc},
		CipherTextLenXor: 0x5678,
		CipherText:       []byte("test"),
	}

	got1, err := ol.HeaderBytes()
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	got2, err := ol.HeaderBytes()
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if !bytes.Equal(got1, got2) {
		t.Fatalf("HeaderBytes() is not idempotent")
	}
}

func TestOnionLayer_Bytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ol          onion.OnionLayer
		wantLen     int
		checkHeader bool
	}{
		{
			name: "with ciphertext",
			ol: onion.OnionLayer{
				EPK:              [32]byte{0x01, 0x02, 0x03},
				WrappedKeys:      [3]onion.WrappedKey{},
				Flags:            0x00,
				PayloadNonce:     [12]byte{0xaa, 0xbb, 0xcc},
				CipherTextLenXor: 0x1234,
				CipherText:       []byte("Hello World"),
			},
			wantLen:     onion.FixedHeaderSize + 11,
			checkHeader: true,
		},
		{
			name: "empty ciphertext",
			ol: onion.OnionLayer{
				EPK:              [32]byte{},
				WrappedKeys:      [3]onion.WrappedKey{},
				Flags:            0x00,
				PayloadNonce:     [12]byte{},
				CipherTextLenXor: 0x0000,
				CipherText:       []byte{},
			},
			wantLen:     onion.FixedHeaderSize,
			checkHeader: true,
		},
		{
			name: "nil ciphertext",
			ol: onion.OnionLayer{
				EPK:              [32]byte{},
				WrappedKeys:      [3]onion.WrappedKey{},
				Flags:            0x00,
				PayloadNonce:     [12]byte{},
				CipherTextLenXor: 0x0000,
				CipherText:       nil,
			},
			wantLen:     onion.FixedHeaderSize,
			checkHeader: false,
		},
		{
			name: "large ciphertext",
			ol: onion.OnionLayer{
				EPK:              [32]byte{0xff},
				WrappedKeys:      [3]onion.WrappedKey{},
				Flags:            0xff,
				PayloadNonce:     [12]byte{0xff},
				CipherTextLenXor: 0xffff,
				CipherText:       make([]byte, 4096),
			},
			wantLen:     onion.FixedHeaderSize + 4096,
			checkHeader: true,
		},
		{
			name: "single byte ciphertext",
			ol: onion.OnionLayer{
				EPK:              [32]byte{0x42},
				WrappedKeys:      [3]onion.WrappedKey{},
				Flags:            0x01,
				PayloadNonce:     [12]byte{0x11},
				CipherTextLenXor: 0x0001,
				CipherText:       []byte{0xaa},
			},
			wantLen:     onion.FixedHeaderSize + 1,
			checkHeader: true,
		},
		{
			name: "full wrapped keys with ciphertext",
			ol: onion.OnionLayer{
				EPK: [32]byte{0x01, 0x02, 0x03, 0x04},
				WrappedKeys: [3]onion.WrappedKey{
					{
						Nonce:      [12]byte{0xa1, 0xa2},
						CipherText: [64]byte{0xb1, 0xb2},
					},
					{
						Nonce:      [12]byte{0xc1, 0xc2},
						CipherText: [64]byte{0xd1, 0xd2},
					},
					{
						Nonce:      [12]byte{0xe1, 0xe2},
						CipherText: [64]byte{0xf1, 0xf2},
					},
				},
				Flags:            0x42,
				PayloadNonce:     [12]byte{0xcc, 0xdd},
				CipherTextLenXor: 0xabcd,
				CipherText:       []byte{0x11, 0x22, 0x33, 0x44},
			},
			wantLen:     onion.FixedHeaderSize + 4,
			checkHeader: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.ol.Bytes()
			if err != nil {
				t.Fatalf("Bytes() unexpected error: %v", err)
			}

			if len(got) != tt.wantLen {
				t.Fatalf("length mismatch:\n\tgot:  %d\n\twant: %d",
					len(got), tt.wantLen)
			}

			if tt.checkHeader {
				header, err := tt.ol.HeaderBytes()
				if err != nil {
					t.Fatalf("HeaderBytes() failed: %v", err)
				}

				if !bytes.Equal(got[:len(header)], header) {
					t.Fatalf("header portion does not match HeaderBytes()")
				}

				if len(tt.ol.CipherText) > 0 {
					cipherTextStart := len(header)
					if !bytes.Equal(got[cipherTextStart:], tt.ol.CipherText) {
						t.Fatalf("ciphertext mismatch:\n\tgot:  %x\n\twant: %x",
							got[cipherTextStart:], tt.ol.CipherText)
					}
				}
			}
		})
	}
}

func TestOnionLayer_Bytes_Idempotent(t *testing.T) {
	t.Parallel()

	ol := onion.OnionLayer{
		EPK:              [32]byte{0x01},
		WrappedKeys:      [3]onion.WrappedKey{},
		Flags:            0x99,
		PayloadNonce:     [12]byte{0xaa},
		CipherTextLenXor: 0x1111,
		CipherText:       []byte("test"),
	}

	got1, err := ol.Bytes()
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	got2, err := ol.Bytes()
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if !bytes.Equal(got1, got2) {
		t.Fatalf("Bytes() is not idempotent")
	}
}

func TestOnionLayer_BytesPadded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ol          onion.OnionLayer
		wantErr     bool
		errContains string
		checkSize   bool
		expectSize  int
	}{
		{
			name: "small packet gets padded",
			ol: onion.OnionLayer{
				EPK:              [32]byte{0x01, 0x02, 0x03},
				WrappedKeys:      [3]onion.WrappedKey{},
				Flags:            0x00,
				PayloadNonce:     [12]byte{0xaa, 0xbb, 0xcc},
				CipherTextLenXor: 0x1234,
				CipherText:       []byte("Hello"),
			},
			wantErr:    false,
			checkSize:  true,
			expectSize: onion.PacketSize,
		},
		{
			name: "empty ciphertext gets padded",
			ol: onion.OnionLayer{
				EPK:              [32]byte{},
				WrappedKeys:      [3]onion.WrappedKey{},
				Flags:            0x00,
				PayloadNonce:     [12]byte{},
				CipherTextLenXor: 0x0000,
				CipherText:       []byte{},
			},
			wantErr:    false,
			checkSize:  true,
			expectSize: onion.PacketSize,
		},
		{
			name: "large ciphertext gets padded",
			ol: onion.OnionLayer{
				EPK:              [32]byte{0xff},
				WrappedKeys:      [3]onion.WrappedKey{},
				Flags:            0xff,
				PayloadNonce:     [12]byte{0xff},
				CipherTextLenXor: 0xffff,
				CipherText:       make([]byte, 2048),
			},
			wantErr:    false,
			checkSize:  true,
			expectSize: onion.PacketSize,
		},
		{
			name: "packet at exact size",
			ol: onion.OnionLayer{
				EPK:              [32]byte{},
				WrappedKeys:      [3]onion.WrappedKey{},
				Flags:            0x00,
				PayloadNonce:     [12]byte{},
				CipherTextLenXor: 0x0000,
				CipherText:       make([]byte, onion.PacketSize-onion.FixedHeaderSize),
			},
			wantErr:    false,
			checkSize:  true,
			expectSize: onion.PacketSize,
		},
		{
			name: "packet overflow returns error",
			ol: onion.OnionLayer{
				EPK:              [32]byte{},
				WrappedKeys:      [3]onion.WrappedKey{},
				Flags:            0x00,
				PayloadNonce:     [12]byte{},
				CipherTextLenXor: 0x0000,
				CipherText:       make([]byte, onion.PacketSize-onion.FixedHeaderSize+1),
			},
			wantErr:     true,
			errContains: "packet overflow",
			checkSize:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.ol.BytesPadded()

			if (err != nil) != tt.wantErr {
				t.Fatalf("BytesPadded() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("error message mismatch:\n\tgot:  %v\n\twant to contain: %s",
						err, tt.errContains)
				}
				return
			}

			if tt.checkSize && len(got) != tt.expectSize {
				t.Fatalf("size mismatch:\n\tgot:  %d\n\twant: %d",
					len(got), tt.expectSize)
			}
		})
	}
}

func TestOnionLayer_BytesPadded_ContainsOriginalData(t *testing.T) {
	t.Parallel()

	ol := onion.OnionLayer{
		EPK:              [32]byte{0x01, 0x02, 0x03},
		WrappedKeys:      [3]onion.WrappedKey{},
		Flags:            0x42,
		PayloadNonce:     [12]byte{0xaa, 0xbb, 0xcc},
		CipherTextLenXor: 0x1234,
		CipherText:       []byte("test payload"),
	}

	rawBytes, err := ol.Bytes()
	if err != nil {
		t.Fatalf("Bytes() failed: %v", err)
	}

	padded, err := ol.BytesPadded()
	if err != nil {
		t.Fatalf("BytesPadded() failed: %v", err)
	}

	if !bytes.Equal(padded[:len(rawBytes)], rawBytes) {
		t.Fatalf("padded packet does not contain original data at start")
	}
}

func TestOnionLayer_BytesPadded_PaddingIsRandom(t *testing.T) {
	t.Parallel()

	ol := onion.OnionLayer{
		EPK:              [32]byte{0x01},
		WrappedKeys:      [3]onion.WrappedKey{},
		Flags:            0x00,
		PayloadNonce:     [12]byte{0xaa},
		CipherTextLenXor: 0x0001,
		CipherText:       []byte("small"),
	}

	padded1, err := ol.BytesPadded()
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	padded2, err := ol.BytesPadded()
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	rawBytes, _ := ol.Bytes()
	paddingStart := len(rawBytes)

	padding1 := padded1[paddingStart:]
	padding2 := padded2[paddingStart:]

	if bytes.Equal(padding1, padding2) {
		t.Fatalf("padding should be random, got identical padding")
	}
}

func TestOnionLayer_BytesPadded_NoPaddingNeeded(t *testing.T) {
	t.Parallel()

	ol := onion.OnionLayer{
		EPK:              [32]byte{},
		WrappedKeys:      [3]onion.WrappedKey{},
		Flags:            0x00,
		PayloadNonce:     [12]byte{},
		CipherTextLenXor: 0x0000,
		CipherText:       make([]byte, onion.PacketSize-onion.FixedHeaderSize),
	}

	got, err := ol.BytesPadded()
	if err != nil {
		t.Fatalf("BytesPadded() failed: %v", err)
	}

	if len(got) != onion.PacketSize {
		t.Fatalf("size mismatch:\n\tgot:  %d\n\twant: %d",
			len(got), onion.PacketSize)
	}
}

func TestOnionLayer_BytesPadded_NotIdempotent(t *testing.T) {
	t.Parallel()

	ol := onion.OnionLayer{
		EPK:              [32]byte{0x01},
		WrappedKeys:      [3]onion.WrappedKey{},
		Flags:            0x99,
		PayloadNonce:     [12]byte{0xaa},
		CipherTextLenXor: 0x1111,
		CipherText:       []byte("test"),
	}

	got1, err := ol.BytesPadded()
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	got2, err := ol.BytesPadded()
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	rawBytes, _ := ol.Bytes()
	dataLen := len(rawBytes)

	if !bytes.Equal(got1[:dataLen], got2[:dataLen]) {
		t.Fatalf("data portion should be identical across calls")
	}
}

func TestOnionLayer_Parse(t *testing.T) {
	t.Parallel()

	flagsOff := 32 + onion.MaxWrappedKey*onion.WrappedKeySize
	nonceOff := flagsOff + 1
	lenOff := nonceOff + 12

	tests := []struct {
		name          string
		data          []byte
		wantErr       bool
		errContains   string
		checkEPK      bool
		expectedEPK   [32]byte
		checkFlags    bool
		expectedFlags uint8

		checkNonce    bool
		expectedNonce [12]byte

		checkLenXor    bool
		expectedLenXor uint16

		checkWrapped0 bool
		w0Nonce0      byte
		w0CTLast      byte

		checkCipherText bool
		expectedCT      []byte
	}{
		{
			name: "valid minimal packet",
			data: func() []byte {
				buf := make([]byte, onion.FixedHeaderSize)

				buf[0] = 0x01
				buf[31] = 0x1f

				buf[32] = 0x99

				buf[32+12+63] = 0x77

				buf[flagsOff] = 0x42

				buf[nonceOff] = 0xaa
				buf[nonceOff+11] = 0xbb

				buf[lenOff] = 0x12
				buf[lenOff+1] = 0x34

				return buf
			}(),
			wantErr:       false,
			checkEPK:      true,
			expectedEPK:   [32]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1f},
			checkFlags:    true,
			expectedFlags: 0x42,

			checkNonce:    true,
			expectedNonce: [12]byte{0xaa, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xbb},

			checkLenXor:    true,
			expectedLenXor: 0x1234,

			checkWrapped0: true,
			w0Nonce0:      0x99,
			w0CTLast:      0x77,
		},
		{
			name: "packet with ciphertext",
			data: func() []byte {
				payload := []byte("test data!")
				buf := make([]byte, onion.FixedHeaderSize+len(payload))
				copy(buf[onion.FixedHeaderSize:], payload)
				return buf
			}(),
			wantErr:         false,
			checkCipherText: true,
			expectedCT:      []byte("test data!"),
		},
		{
			name:           "all zero packet",
			data:           make([]byte, onion.FixedHeaderSize),
			wantErr:        false,
			checkEPK:       true,
			expectedEPK:    [32]byte{},
			checkFlags:     true,
			expectedFlags:  0x00,
			checkNonce:     true,
			expectedNonce:  [12]byte{},
			checkLenXor:    true,
			expectedLenXor: 0x0000,
		},
		{
			name: "all max values packet",
			data: func() []byte {
				buf := make([]byte, onion.FixedHeaderSize)
				for i := range buf {
					buf[i] = 0xff
				}
				return buf
			}(),
			wantErr:        false,
			checkFlags:     true,
			expectedFlags:  0xff,
			checkNonce:     true,
			expectedNonce:  [12]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			checkLenXor:    true,
			expectedLenXor: 0xffff,
		},
		{
			name:        "data too short by 1 byte",
			data:        make([]byte, onion.FixedHeaderSize-1),
			wantErr:     true,
			errContains: "data too short",
		},
		{
			name:        "empty data",
			data:        []byte{},
			wantErr:     true,
			errContains: "data too short",
		},
		{
			name:        "nil data",
			data:        nil,
			wantErr:     true,
			errContains: "data too short",
		},
		{
			name: "large packet with ciphertext",
			data: func() []byte {
				buf := make([]byte, onion.PacketSize)
				buf[0] = 0xaa
				buf[31] = 0xbb

				ct := buf[onion.FixedHeaderSize:]
				ct[0] = 0xDE
				ct[len(ct)-1] = 0xAD

				return buf
			}(),
			wantErr:         false,
			checkEPK:        true,
			expectedEPK:     [32]byte{0xaa, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xbb},
			checkCipherText: true,
			expectedCT: func() []byte {
				ct := make([]byte, onion.PacketSize-onion.FixedHeaderSize)
				ct[0] = 0xDE
				ct[len(ct)-1] = 0xAD
				return ct
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var ol onion.OnionLayer
			err := ol.Parse(tt.data)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
					t.Fatalf("error message mismatch:\n\tgot:  %v\n\twant to contain: %s", err, tt.errContains)
				}
				return
			}

			if tt.checkEPK && ol.EPK != tt.expectedEPK {
				t.Fatalf("Parse(): EPK mismatch:\n\tgot:  %x\n\twant: %x", ol.EPK, tt.expectedEPK)
			}

			if tt.checkFlags && ol.Flags != tt.expectedFlags {
				t.Fatalf("Parse(): Flags mismatch:\n\tgot:  0x%02x\n\twant: 0x%02x", ol.Flags, tt.expectedFlags)
			}

			if tt.checkNonce && ol.PayloadNonce != tt.expectedNonce {
				t.Fatalf("Parse(): PayloadNonce mismatch:\n\tgot:  %x\n\twant: %x", ol.PayloadNonce, tt.expectedNonce)
			}

			if tt.checkLenXor && ol.CipherTextLenXor != tt.expectedLenXor {
				t.Fatalf("Parse(): CipherTextLenXor mismatch:\n\tgot:  0x%04x\n\twant: 0x%04x", ol.CipherTextLenXor, tt.expectedLenXor)
			}

			if tt.checkWrapped0 {
				if ol.WrappedKeys[0].Nonce[0] != tt.w0Nonce0 {
					t.Fatalf("Parse(): WrappedKeys[0].Nonce[0] mismatch:\n\tgot:  0x%02x\n\twant: 0x%02x",
						ol.WrappedKeys[0].Nonce[0], tt.w0Nonce0)
				}
				if ol.WrappedKeys[0].CipherText[63] != tt.w0CTLast {
					t.Fatalf("Parse(): WrappedKeys[0].CipherText[63] mismatch:\n\tgot:  0x%02x\n\twant: 0x%02x",
						ol.WrappedKeys[0].CipherText[63], tt.w0CTLast)
				}
			}

			wantCTLen := len(tt.data) - onion.FixedHeaderSize
			if len(ol.CipherText) != wantCTLen {
				t.Fatalf("Parse(): CipherText length mismatch:\n\tgot:  %d\n\twant: %d", len(ol.CipherText), wantCTLen)
			}

			if tt.checkCipherText && !bytes.Equal(ol.CipherText, tt.expectedCT) {
				t.Fatalf("Parse(): CipherText content mismatch:\n\tgot:  %x\n\twant: %x", ol.CipherText, tt.expectedCT)
			}
		})
	}
}

func TestOnionLayer_CipherTextLen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		cipherKey     [32]byte
		payloadNonce  [12]byte
		cipherTextLen int
		wantLen       uint16
	}{
		{
			name: "len matches ciphertext (non-zero)",
			cipherKey: [32]byte{
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
				0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
				0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
				0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			},
			payloadNonce:  [12]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			cipherTextLen: 100,
			wantLen:       100,
		},
		{
			name:          "len matches ciphertext (zero)",
			cipherKey:     [32]byte{},
			payloadNonce:  [12]byte{},
			cipherTextLen: 0,
			wantLen:       0,
		},
		{
			name: "len matches ciphertext (bigger)",
			cipherKey: [32]byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			},
			payloadNonce:  [12]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			cipherTextLen: 1000,
			wantLen:       1000,
		},
		{
			name: "returned len is the encoded one (not ciphertext len)",
			cipherKey: [32]byte{
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
				0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
				0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
				0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			},
			payloadNonce:  [12]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			cipherTextLen: 10,
			wantLen:       49,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ol := onion.OnionLayer{
				PayloadNonce: tt.payloadNonce,
				CipherText:   make([]byte, tt.cipherTextLen),
			}

			mask, err := onion.CipherTextLenMask16(tt.cipherKey, tt.payloadNonce)
			if err != nil {
				t.Fatalf("CipherTextLenMask16() failed: %v", err)
			}

			ol.CipherTextLenXor = tt.wantLen ^ mask

			got, err := ol.CipherTextLen(tt.cipherKey)
			if err != nil {
				t.Fatalf("CipherTextLen() returned error: %v", err)
			}

			if got != int(tt.wantLen) {
				t.Fatalf("CipherTextLen() mismatch:\n\tgot:  %d\n\twant: %d", got, tt.wantLen)
			}
		})
	}
}

func TestOnionLayer_TrimCipherText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		nonce       [12]byte
		key         [32]byte
		ctLen       int
		realLen     uint16
		wantErr     bool
		errContains string
	}{
		{
			name:    "trims ciphertext to realLen",
			nonce:   [12]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			key:     [32]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f},
			ctLen:   100,
			realLen: 49,
			wantErr: false,
		},
		{
			name:        "realLen exceeds ciphertext length -> error",
			nonce:       [12]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			key:         [32]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f},
			ctLen:       10,
			realLen:     49,
			wantErr:     true,
			errContains: "invalid cipher text length",
		},
		{
			name:    "zero realLen keeps empty slice",
			nonce:   [12]byte{},
			key:     [32]byte{},
			ctLen:   0,
			realLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ol := onion.OnionLayer{
				PayloadNonce: tt.nonce,
				CipherText:   make([]byte, tt.ctLen),
			}

			for i := range ol.CipherText {
				ol.CipherText[i] = byte(i)
			}

			mask, err := onion.CipherTextLenMask16(tt.key, tt.nonce)
			if err != nil {
				t.Fatalf("CipherTextLenMask16() failed: %v", err)
			}

			ol.CipherTextLenXor = tt.realLen ^ mask

			orig := append([]byte(nil), ol.CipherText...)

			err = ol.TrimCipherText(tt.key)
			if (err != nil) != tt.wantErr {
				t.Fatalf("TrimCipherText() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
					t.Fatalf("error message mismatch:\n\tgot:  %v\n\twant to contain: %s", err, tt.errContains)
				}
				return
			}

			if len(ol.CipherText) != int(tt.realLen) {
				t.Fatalf("trimmed length mismatch:\n\tgot:  %d\n\twant: %d", len(ol.CipherText), tt.realLen)
			}

			if !bytes.Equal(ol.CipherText, orig[:tt.realLen]) {
				t.Fatalf("trimmed content mismatch:\n\tgot:  %x\n\twant: %x", ol.CipherText, orig[:tt.realLen])
			}
		})
	}
}

// -------------------------
// Benchmark functions
// -------------------------

// Helper to build Layer for Benchmark
func benchLayer(cipherLen int) onion.OnionLayer {
	var ol onion.OnionLayer
	for i := range 32 {
		ol.EPK[i] = byte(i)
	}
	for i := range onion.MaxWrappedKey {
		for j := range 12 {
			ol.WrappedKeys[i].Nonce[j] = byte(i + j)
		}
		for j := range 64 {
			ol.WrappedKeys[i].CipherText[j] = byte(i ^ j)
		}
	}

	ol.Flags = 0x42
	for i := range 12 {
		ol.PayloadNonce[i] = byte(0xaa + i)
	}

	ol.CipherTextLenXor = 0x1234
	ol.CipherText = make([]byte, cipherLen)
	for i := range ol.CipherText {
		ol.CipherText[i] = byte(i)
	}

	return ol
}

func BenchmarkOnionLayer_HeaderBytes(b *testing.B) {
	ol := benchLayer(0)
	b.ReportAllocs()

	for b.Loop() {
		_, _ = ol.HeaderBytes()
	}
}

func BenchmarkOnionLayer_Bytes_Sizes(b *testing.B) {
	cases := []struct {
		name string
		size int
	}{
		{name: "0", size: 0},
		{name: "32", size: 32},
		{name: "256", size: 256},
		{name: "1024", size: 1024},
		{name: "4096", size: 4096},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			ol := benchLayer(tc.size)
			b.ReportAllocs()
			b.SetBytes(int64(onion.FixedHeaderSize + tc.size))

			for b.Loop() {
				_, _ = ol.Bytes()
			}
		})
	}
}

func BenchmarkOnionLayer_BytesPadded(b *testing.B) {
	ol := benchLayer(512)
	b.ReportAllocs()
	b.SetBytes(int64(onion.PacketSize))

	for b.Loop() {
		_, _ = ol.BytesPadded()
	}
}

func BenchmarkOnionLayer_Parse(b *testing.B) {
	ol := benchLayer(512)
	raw, err := ol.Bytes()
	if err != nil {
		b.Fatalf("Bytes() failed: %v", err)
	}

	b.ReportAllocs()
	b.SetBytes(int64(len(raw)))

	for b.Loop() {
		var dst onion.OnionLayer
		_ = dst.Parse(raw)
	}
}

func BenchmarkCipherTextLenMask16(b *testing.B) {
	var key [32]byte
	var nonce [12]byte

	for i := range 32 {
		key[i] = byte(i)
	}
	for i := range 12 {
		nonce[i] = byte(0xaa + i)
	}

	b.ReportAllocs()

	for b.Loop() {
		_, _ = onion.CipherTextLenMask16(key, nonce)
	}
}

func BenchmarkOnionLayer_TrimCipherText(b *testing.B) {
	var key [32]byte
	for i := range 32 {
		key[i] = byte(i)
	}

	ol := benchLayer(2048)

	mask, err := onion.CipherTextLenMask16(key, ol.PayloadNonce)
	if err != nil {
		b.Fatalf("mask failed: %v", err)
	}

	ol.CipherTextLenXor = uint16(512) ^ mask

	b.ReportAllocs()

	for b.Loop() {
		tmp := ol
		tmp.CipherText = append([]byte(nil), ol.CipherText...)
		_ = tmp.TrimCipherText(key)
	}
}
