//go:build linux
// +build linux

package cmd

import (
	"fmt"

	"ctrz/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ctrz version: \t%s\n", version.Version)
		fmt.Printf("Commit: \t%s\n", version.Commit)
		fmt.Printf("Build date: \t%s\n", version.BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
