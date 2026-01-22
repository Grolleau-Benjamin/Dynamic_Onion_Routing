package onion_test

import (
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/onion"
)

func TestFlags_IsLastServer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flags    uint8
		expected bool
	}{
		{
			name:     "Is last server",
			flags:    0x08,
			expected: true,
		},
		{
			name:     "Is last server with complex flag",
			flags:    0xff,
			expected: true,
		},
		{
			name:     "Is not last server",
			flags:    0x00,
			expected: false,
		},
		{
			name:     "Is not last server with complex flag",
			flags:    0xf7,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := onion.IsLastServer(tt.flags)
			if got != tt.expected {
				t.Fatalf("IsLastServer() unexpected result.\n\tgot: %v\n\twant: %v", got, tt.expected)
			}
		})
	}
}

func TestFlags_GetNbNextHops(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flags    uint8
		expected uint8
	}{
		{
			name:     "0 next hops",
			flags:    0x00,
			expected: 0,
		},
		{
			name:     "0 next hops - complex flag",
			flags:    0xf8,
			expected: 0,
		},
		{
			name:     "7 next hops (Max)",
			flags:    0x07,
			expected: 7,
		},
		{
			name:     "5 next hops ",
			flags:    0x05,
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := onion.GetNbNextHops(tt.flags)
			if got != tt.expected {
				t.Fatalf("GetNbNextHops() unexpected result.\n\tgot: %d\n\twant: %d", got, tt.expected)
			}
		})
	}
}
