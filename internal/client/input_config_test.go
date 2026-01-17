package client_test

import (
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/client"
)

func TestInputConfig_IsComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ic       client.InputConfig
		expected bool
	}{
		{
			name: "valid input data",
			ic: client.InputConfig{
				OnionPath: "[::1]:62503",
				Dest:      "[::1]:8080",
				Payload:   "datas",
			},
			expected: true,
		},
		{
			name: "missing Onion Path",
			ic: client.InputConfig{
				OnionPath: "",
				Dest:      "[::1]:8080",
				Payload:   "datas",
			},
			expected: false,
		},
		{
			name: "missing Dest",
			ic: client.InputConfig{
				OnionPath: "[::1]:62503",
				Dest:      "",
				Payload:   "datas",
			},
			expected: false,
		},
		{
			name: "missing Payload",
			ic: client.InputConfig{
				OnionPath: "[::1]:62503",
				Dest:      "[::1]:8080",
				Payload:   "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.ic.IsComplete()
			if got != tt.expected {
				t.Fatalf("IsComplete() mismatch: \ngot: %v\nwant: %v", got, tt.expected)
			}
		})
	}
}

func TestInputConfig_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ic       client.InputConfig
		expected string
	}{
		{
			name: "valid input data",
			ic: client.InputConfig{
				OnionPath: "[::1]:62503",
				Dest:      "[::1]:8080",
				Payload:   "datas",
			},
			expected: "InputConfig{\n\tOnionPath: [::1]:62503,\n\tDest: [::1]:8080,\n\tPayload: datas\n}",
		},
		{
			name: "missing Onion Path",
			ic: client.InputConfig{
				OnionPath: "",
				Dest:      "[::1]:8080",
				Payload:   "datas",
			},
			expected: "InputConfig{\n\tOnionPath: ,\n\tDest: [::1]:8080,\n\tPayload: datas\n}",
		},
		{
			name: "missing Dest",
			ic: client.InputConfig{
				OnionPath: "[::1]:62503",
				Dest:      "",
				Payload:   "datas",
			},
			expected: "InputConfig{\n\tOnionPath: [::1]:62503,\n\tDest: ,\n\tPayload: datas\n}",
		},
		{
			name: "missing Payload",
			ic: client.InputConfig{
				OnionPath: "[::1]:62503",
				Dest:      "[::1]:8080",
				Payload:   "",
			},
			expected: "InputConfig{\n\tOnionPath: [::1]:62503,\n\tDest: [::1]:8080,\n\tPayload: \n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.ic.String()
			if got != tt.expected {
				t.Fatalf("IsComplete() mismatch: \ngot: %v\nwant: %v", got, tt.expected)
			}
		})
	}
}
