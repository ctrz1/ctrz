package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ctrz",
	Short: "Find something later",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run `ctrz help` for available commands")
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
