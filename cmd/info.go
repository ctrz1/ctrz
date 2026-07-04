package cmd

import (
	"ctrz/misc"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:                "info <container name>",
	Short:              "Print the stored information of a container",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
			os.Exit(0)
		}
		containerInfo, err := misc.GetRawContainerDataFromName(args[0])
		if err != nil {
			log.Fatalf("Error retrieving container info: %v\n", err)
		}
		fmt.Printf("%s\n", string(containerInfo))
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
