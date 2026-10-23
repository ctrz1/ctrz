//go:build linux

package cmd

import (
	"ctrz/runtime"
	"log"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean all files, networking rules and devices created by ctrz",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		r, err := runtime.New()
		if err != nil {
			log.Fatal(err)
		}
		if err := r.Clean(); err != nil {
			log.Fatalf("Error cleaning system: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
