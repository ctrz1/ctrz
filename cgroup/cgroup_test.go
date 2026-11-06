//go:build linux

package cgroup

import (
	"testing"
)

func TestParseCPUMax(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantQuota  int64
		wantPeriod uint64
		wantErr    bool
	}{
		{
			name:       "valid CPU limit",
			input:      "50000 100000",
			wantQuota:  50000,
			wantPeriod: 100000,
		},
		{
			name:    "missing quota",
			input:   "100000",
			wantErr: true,
		},
		{
			name:    "too many fields",
			input:   "50000 100000 200000",
			wantErr: true,
		},
		{
			name:    "invalid quota",
			input:   "abc 100000",
			wantErr: true,
		},
		{
			name:    "negative quota",
			input:   "-50000 100000",
			wantErr: true,
		},
		{
			name:    "zero quota",
			input:   "0 100000",
			wantErr: true,
		},
		{
			name:    "invalid period",
			input:   "50000 abc",
			wantErr: true,
		},
		{
			name:    "zero period",
			input:   "50000 0",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace",
			input:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quota, period, err := parseCPUMax(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if quota != tt.wantQuota {
				t.Errorf("Got quota = %d, want %d",
					quota, tt.wantQuota)
			}

			if period != tt.wantPeriod {
				t.Errorf("Got period = %d, want %d",
					period, tt.wantPeriod)
			}
		})
	}
}
