package crypto_test

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/crypto"
)

func TestChacha20poly1305_ChachaEncrypt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		key       [32]byte
		nonce     [12]byte
		plaintext []byte
		aad       []byte
		expected  []byte
		wantErr   bool
	}{
		{
			name: "normal case",
			key: [32]byte{
				0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE, 0xBA, 0xBE,
			},
			nonce:     [12]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			plaintext: []byte("DOR_Protocol"),
			aad:       []byte("Niconico"),
			expected:  []byte{92, 39, 112, 230, 91, 9, 10, 90, 47, 28, 158, 208, 141, 64, 30, 65, 47, 13, 59, 117, 103, 88, 132, 249, 44, 92, 97, 82},
			wantErr:   false,
		},
		{
			name: "empty plaintext and aad",
			key: [32]byte{
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			},
			nonce:     [12]byte{6, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			plaintext: []byte{},
			aad:       []byte{},
			expected:  []byte{105, 20, 95, 75, 246, 164, 127, 143, 101, 120, 232, 229, 213, 55, 193, 162},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ciphertext, err := crypto.ChachaEncrypt(tt.key, tt.nonce, tt.plaintext, tt.aad)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ChachaEncrypt() unexpected error = %v", err)
			}
			if !bytes.Equal(ciphertext, tt.expected) {
				t.Fatalf("ciphertext mismatch: \n\tgot: %v \n\twant: %v", ciphertext, tt.expected)
			}
		})
	}
}

func TestChacha20poly1305_ChachaDecrypt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		key        [32]byte
		nonce      [12]byte
		ciphertext []byte
		aad        []byte
		expected   []byte
		wantErr    bool
	}{
		{
			name: "normal case",
			key: [32]byte{
				0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE, 0xBA, 0xBE,
			},
			nonce:      [12]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			ciphertext: []byte{92, 39, 112, 230, 91, 9, 10, 90, 47, 28, 158, 208, 141, 64, 30, 65, 47, 13, 59, 117, 103, 88, 132, 249, 44, 92, 97, 82},
			aad:        []byte("Niconico"),
			expected:   []byte("DOR_Protocol"),
			wantErr:    false,
		},
		{
			name: "empty ciphertext and aad",
			key: [32]byte{
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			},
			nonce:      [12]byte{6, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			ciphertext: []byte{},
			aad:        []byte{},
			expected:   []byte{},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plaintext, err := crypto.ChachaDecrypt(tt.key, tt.nonce, tt.ciphertext, tt.aad)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ChachaDecrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !bytes.Equal(plaintext, tt.expected) {
				t.Fatalf("plaintext mismatch \n\tgot: %v\n\twant: %v", plaintext, tt.expected)
			}
		})
	}
}

func TestChacha20poly1305_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		key       [32]byte
		plaintext []byte
		aad       []byte
	}{
		{
			name: "round trip with normal data",
			key: [32]byte{
				0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE, 0xBA, 0xBE,
			},
			plaintext: []byte("DOR_Protocol"),
			aad:       []byte("03312003"),
		},
		{
			name: "round trip with empty data",
			key: [32]byte{
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			},
			plaintext: []byte{},
			aad:       []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var nonce [12]byte
			if _, err := rand.Read(nonce[:]); err != nil {
				t.Fatalf("failed to generate random nonce: %v", err)
			}

			ciphertext, err := crypto.ChachaEncrypt(tt.key, nonce, tt.plaintext, tt.aad)
			if err != nil {
				t.Fatalf("ChachaEncrypt() error: %v", err)
			}

			decrypted, err := crypto.ChachaDecrypt(tt.key, nonce, ciphertext, tt.aad)
			if err != nil {
				t.Fatalf("ChachaDecrypt() error: %v", err)
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Fatalf("round trip failed: \n\tgot: %v\n\t want: %v", decrypted, tt.plaintext)
			}
		})
	}
}
