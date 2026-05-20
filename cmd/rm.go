package cmd

import (
	"ctrz/misc"
	"log"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use: "rm",
	Short: "Remove a container",
	Run: func(cmd *cobra.Command, args []string)  {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatalf("unable to retrieve name: %v", err)
		}
		forceKill, err := cmd.Flags().GetBool("force")
		if err != nil {
			log.Fatalf("unable to retrieve force: %v", err)
		}
		if err := misc.RemoveContainerByName(name, forceKill); err != nil {
			log.Fatalf("Error cleaning up container: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().String("name", "", "Name of the container to be removed")
	rmCmd.Flags().BoolP("force", "9", false, "Uses KILL instead TERM command. This forces a process to stop immediately and does not allow for cleanup routines")
}