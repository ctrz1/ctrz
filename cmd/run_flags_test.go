//go:build linux

package cmd

import (
	"testing"

	"ctrz/config"
)

func TestRunFlagsReachCLIOptions(t *testing.T) {
	flags := runCmd.Flags()

	base, _ := getFlags(runCmd)
	if base.Name != nil || base.Detach != nil || base.Remove != nil ||
		base.Ports != nil || base.Environment != nil || base.CPU != nil {
		t.Fatalf("no flags were set but getFlags reported some: %+v", base)
	}

	tests := []struct {
		flag  string
		value string
		set   func(config.CLIOptions) bool
	}{
		{"name", "x", func(o config.CLIOptions) bool { return o.Name != nil && *o.Name == "x" }},
		{"detach", "true", func(o config.CLIOptions) bool { return o.Detach != nil && *o.Detach }},
		{"rm", "true", func(o config.CLIOptions) bool { return o.Remove != nil && *o.Remove }},
		{"port", "8080:80", func(o config.CLIOptions) bool { return o.Ports != nil && len(*o.Ports) == 1 }},
		{"env", "A=b", func(o config.CLIOptions) bool { return o.Environment != nil && len(*o.Environment) == 1 }},
		{"cpu", "50", func(o config.CLIOptions) bool { return o.CPU != nil && *o.CPU == "50000 100000" }},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Fatalf("run has no --%s flag", tt.flag)
			}
			t.Cleanup(func() { f.Changed = false })

			if err := flags.Set(tt.flag, tt.value); err != nil {
				t.Fatalf("setting --%s=%s: %v", tt.flag, tt.value, err)
			}
			got, _ := getFlags(runCmd)
			if !tt.set(got) {
				t.Fatalf("--%s=%s was given but getFlags() did not report it: %+v", tt.flag, tt.value, got)
			}
		})
	}
}
