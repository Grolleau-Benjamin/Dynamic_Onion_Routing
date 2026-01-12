package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

type chachaTestCase struct {
	name      string
	keyHex    string
	nonce     [12]byte
	plaintext []byte
	aad       []byte
	tamper    bool
	wantErr   bool
}

func getChachaCases() []chachaTestCase {
	return []chachaTestCase{
		{
			name:      "encrypt-decrypt",
			keyHex:    "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f",
			nonce:     [12]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			plaintext: []byte("DOR - Dynamic Onion Routing"),
			aad:       []byte("app-aad"),
			tamper:    false,
			wantErr:   false,
		},
		{
			name:      "tamper-detection",
			keyHex:    "",
			nonce:     [12]byte{},
			plaintext: []byte("DOR - Dynamic Onion Routing"),
			aad:       nil,
			tamper:    true,
			wantErr:   true,
		},
		{
			name:      "empty-plaintext",
			keyHex:    "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f",
			nonce:     [12]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c},
			plaintext: []byte{},
			aad:       []byte("aad"),
			tamper:    false,
			wantErr:   false,
		},
		{
			name:      "aad-mismatch",
			keyHex:    "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f",
			nonce:     [12]byte{0x09, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x01},
			plaintext: []byte("DOR - Dynamic Onion Routing"),
			aad:       []byte("aad1"),
			tamper:    false,
			wantErr:   true,
		},
		{
			name:      "deterministic",
			keyHex:    "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f",
			nonce:     [12]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b},
			plaintext: []byte("DOR - Dynamic Onion Routing"),
			aad:       nil,
			tamper:    false,
			wantErr:   false,
		},
	}
}

func TestChachaEncrypt(t *testing.T) {
	t.Parallel()
	cases := getChachaCases()

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var key [32]byte
			if tc.keyHex != "" {
				kb, err := hex.DecodeString(tc.keyHex)
				if err != nil {
					t.Fatalf("invalid hex key: %v", err)
				}
				if len(kb) != 32 {
					t.Fatalf("key length != 32: %d", len(kb))
				}
				copy(key[:], kb)
			}

			ciphertext, err := ChachaEncrypt(key, tc.nonce, tc.plaintext, tc.aad)
			if err != nil {
				t.Fatalf("encrypt failed: %v", err)
			}

			if got, want := len(ciphertext), len(tc.plaintext)+Poly1305TagSize; got != want {
				t.Fatalf("ciphertext length mismatch: got %d want %d", got, want)
			}

			if len(tc.plaintext) > 0 && bytes.Equal(ciphertext, tc.plaintext) {
				t.Fatalf("ciphertext equals plaintext")
			}

			if tc.name == "deterministic" {
				c2, err := ChachaEncrypt(key, tc.nonce, tc.plaintext, tc.aad)
				if err != nil {
					t.Fatalf("second encrypt failed: %v", err)
				}
				if !bytes.Equal(ciphertext, c2) {
					t.Fatalf("encryption not deterministic: got %x vs %x", ciphertext, c2)
				}
			}
		})
	}
}

func TestChachaDecrypt(t *testing.T) {
	t.Parallel()
	cases := getChachaCases()

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var key [32]byte
			if tc.keyHex != "" {
				kb, err := hex.DecodeString(tc.keyHex)
				if err != nil {
					t.Fatalf("invalid hex key: %v", err)
				}
				if len(kb) != 32 {
					t.Fatalf("key length != 32: %d", len(kb))
				}
				copy(key[:], kb)
			}

			ciphertext, err := ChachaEncrypt(key, tc.nonce, tc.plaintext, tc.aad)
			if err != nil {
				t.Fatalf("encrypt failed: %v", err)
			}

			cForOpen := make([]byte, len(ciphertext))
			copy(cForOpen, ciphertext)
			if tc.tamper && len(cForOpen) > 0 {
				cForOpen[0] ^= 0xff
			}

			aadToUse := tc.aad
			if tc.name == "aad-mismatch" {
				aadToUse = []byte("aad2")
			}

			pt, err := ChachaDecrypt(key, tc.nonce, cForOpen, aadToUse)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected decrypt error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("decrypt failed: %v", err)
			}

			if !bytes.Equal(pt, tc.plaintext) {
				t.Fatalf("plaintext mismatch: got %x want %x", pt, tc.plaintext)
			}
		})
	}
}
