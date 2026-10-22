//go:build linux
// +build linux

package cmd

import "testing"

func TestConvertBytesToFittingUnit(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{
			name:     "Pass bytes",
			input:    10,
			expected: "10 B",
		},
		{
			name:     "Pass kilobytes",
			input:    1300,
			expected: "1.3 KB",
		},
		{
			name:     "Pass gigabytes",
			input:    5956798763, //5_956_798_763
			expected: "5.5 GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertBytesToFittingUnit(tt.input)
			if got != tt.expected {
				t.Fatalf("got %v, want %v", got, tt.expected)
			}
		})
	}
}
