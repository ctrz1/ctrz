package cmd

import (
	"ctrz/misc"
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ctrz version %s\n", misc.Version)
		fmt.Printf("Commit: %s\n", misc.Commit)
		fmt.Printf("Build date: %s\n", misc.BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
