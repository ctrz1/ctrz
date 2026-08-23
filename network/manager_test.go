//go:build linux

package network

import "testing"

func TestGateway(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		input    string
		wantErr  bool
	}{
		{
			name:     "Parse standard subnet",
			expected: "10.200.1.1/24",
			input:    "10.200.1.0/24",
			wantErr:  false,
		},
		{
			name:     "Parse standard subnet",
			expected: "192.168.1.1/24",
			input:    "192.168.1.0/24",
			wantErr:  false,
		},
		{
			name:     "Parse incorrect subnet",
			expected: "",
			input:    "192.168.1.1.0/24",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gateway(tt.input)
			if err != nil && tt.wantErr == false {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Fatalf("got %v, want %v", got, tt.expected)
			}
		})
	}
}
