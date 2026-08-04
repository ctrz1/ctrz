package network

import (
	"testing"
)

func TestTcpState(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tcp state listening",
			input:    "0A",
			expected: "LISTEN",
		},
		{
			name:     "tcp state etsablished",
			input:    "01",
			expected: "ESTABLISHED",
		},
		{
			name:     "tcp state waiting",
			input:    "06",
			expected: "TIME_WAIT",
		},
		{
			name:     "unknown tcp state",
			input:    "77",
			expected: "77",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tcpState(tt.input)
			if got != tt.expected {
				t.Fatalf("got %s, want %s", got, tt.expected)
			}
		})
	}
}
