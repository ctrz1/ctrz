package cmd

import (
	"fmt"
	"log"
	"ctrz/proc"
	"ctrz/cgroup"

	"github.com/spf13/cobra"
)

var wrapCmd = &cobra.Command{
    Use:   "wrap",
    Short: "'wrap' a process in a debug container",
    Run: func(cmd *cobra.Command, args []string) {
        pid, err := cmd.Flags().GetInt("pid")
		if err != nil {
			log.Fatal(err)
		}
		info, err := proc.Inspect(pid)
		if err != nil {
			log.Fatal(err)			
		}
		fmt.Println(info.String())

		err = cgroup.CreateAndAttach(pid, "20000 100000")
		if err != nil {
			panic(err)
		}
    },
}

func init() {
	rootCmd.AddCommand(wrapCmd)
	wrapCmd.Flags().Int("pid", 0, "Process ID of process that you want to 'wrap' in a container")
	wrapCmd.MarkFlagRequired("pid")
}