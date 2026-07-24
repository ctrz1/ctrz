package network

import (
	"testing"
)

func TestParseProcNetAddr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		ipv6     bool
		expected string
		wantErr  bool
	}{
		{
			name:     "ipv4 basic (loopback style)",
			input:    "0100007F:0035", // 127.0.0.1:53
			ipv6:     false,
			expected: "127.0.0.1:53",
		},
		{
			name:     "ipv4 typical public ip",
			input:    "0C00C80A:20FB", // 10.200.0.12:8443
			ipv6:     false,
			expected: "10.200.0.12:8443",
		},
		{
			name:     "ipv4 all zeros (wildcard)",
			input:    "00000000:1F90", // 0.0.0.0:8080
			ipv6:     false,
			expected: "0.0.0.0:8080",
		},
		{
			name:     "ipv4 max address",
			input:    "FFFFFFFF:01BB", // 255.255.255.255:443
			ipv6:     false,
			expected: "255.255.255.255:443",
		},
		{
			name:    "ipv4 invalid hex length",
			input:   "ABC:1234",
			ipv6:    false,
			wantErr: true,
		},
		{
			name:     "ipv6 loopback",
			input:    "00000000000000000000000000000001:0035",
			ipv6:     true,
			expected: "[::1]:53",
		},
		{
			name:     "ipv6 unspecified address",
			input:    "00000000000000000000000000000000:1F90",
			ipv6:     true,
			expected: "[::]:8080", // net.IP.String() compresses 0s
		},
		{
			name:  "ipv6 typical mixed zero pattern",
			input: "00000000000000000000FFFF0A00000C:20FB",
			ipv6:  true,
			//expected: "[::ffff:10.0.0.12]:8443",
			expected: "[10.0.0.12]:8443", // IPv4 mapped to IPv6
		},
		{
			name:    "ipv6 invalid length",
			input:   "1234:80",
			ipv6:    true,
			wantErr: true,
		},
		{
			name:    "missing separator",
			input:   "0100007F0035",
			ipv6:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProcNetAddr(tt.input, tt.ipv6)
			if err != nil && tt.wantErr == false {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Fatalf("got %q, want %q", got, tt.expected)
			}
		})
	}
}
