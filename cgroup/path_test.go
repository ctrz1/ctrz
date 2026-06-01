package cgroup

import (
	"testing"
)

func TestPathForPID(t *testing.T) {

	tests := []struct {
		name     string
		pid      int
		expected string
		wantErr  bool
	}{
		{
			name:     "basic PID",
			pid:      247,
			expected: "/sys/fs/cgroup/ctrz-247",
		},
		{
			name:     "basic PID",
			pid:      -3,
			expected: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PathForPID(tt.pid)
			if err != nil && tt.wantErr == false {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Fatalf("got %q, want %q", got, tt.expected)
			}
		})
	}

}
