package cmd

import (
	"fmt"
	"log"

	"ctrz/runtime"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <container name>",
	Short: "Print the stored information of a container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerInfo, err := runtime.GetRawContainerDataFromName(args[0])
		if err != nil {
			log.Fatalf("Error retrieving container info: %v\n", err)
		}
		fmt.Printf("%s\n", string(containerInfo))
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
