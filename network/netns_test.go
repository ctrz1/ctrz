//go:build linux

package network

import (
	"ctrz/spec"
	"testing"
)

func TestParsePort(t *testing.T) {
	tests := []struct {
		name     string
		expected spec.PortMapping
		input    string
		wantErr  bool
	}{
		{
			name: "map same ports host -> container",
			expected: spec.PortMapping{
				ContainerPort: 8443,
				HostPort:      8443,
			},
			input:   "8443:8443",
			wantErr: false,
		},
		{
			name: "map different ports host -> container",
			expected: spec.PortMapping{
				ContainerPort: 8080,
				HostPort:      8443,
			},
			input:   "8443:8080",
			wantErr: false,
		},
		{
			name: "map invalid container port",
			expected: spec.PortMapping{
				ContainerPort: 0,
				HostPort:      0,
			},
			input:   "8443:80808",
			wantErr: true,
		},
		{
			name: "map invalid container port",
			expected: spec.PortMapping{
				ContainerPort: 0,
				HostPort:      0,
			},
			input:   "8443:0",
			wantErr: true,
		},
		{
			name: "map invalid host port",
			expected: spec.PortMapping{
				ContainerPort: 0,
				HostPort:      0,
			},
			input:   "84438:8080",
			wantErr: true,
		},
		{
			name: "map invalid host port",
			expected: spec.PortMapping{
				ContainerPort: 0,
				HostPort:      0,
			},
			input:   "0:8080",
			wantErr: true,
		},
		{
			name: "too many ports",
			expected: spec.PortMapping{
				ContainerPort: 0,
				HostPort:      0,
			},
			input:   "8443:8080:8888",
			wantErr: true,
		},
		{
			name: "too few ports",
			expected: spec.PortMapping{
				ContainerPort: 0,
				HostPort:      0,
			},
			input:   "8443",
			wantErr: true,
		},
		{
			name: "negative port numbers",
			expected: spec.PortMapping{
				ContainerPort: 0,
				HostPort:      0,
			},
			input:   "-8443:8080",
			wantErr: true,
		},
		{
			name: "alphanumerical ports",
			expected: spec.PortMapping{
				ContainerPort: 0,
				HostPort:      0,
			},
			input:   "hello:world",
			wantErr: true,
		},
		{
			name: "port number too high",
			expected: spec.PortMapping{
				ContainerPort: 0,
				HostPort:      0,
			},
			input:   "67000:8080",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePorts(tt.input)
			if err != nil && tt.wantErr == false {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ContainerPort != tt.expected.ContainerPort || got.HostPort != tt.expected.HostPort {
				t.Fatalf("got %v, want %v", got, tt.expected)
			}
		})
	}
}
