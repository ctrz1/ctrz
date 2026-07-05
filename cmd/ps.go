package cmd

import (
	"ctrz/misc"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "Print list of (running) containers",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		containers, err := misc.RetrieveAllContainers()
		if err != nil {
			log.Fatalf("Error retrieving containers: %v\n", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		defer w.Flush()

		fmt.Fprintln(w)
		//fmt.Fprintln(w, "-------------------------------------------------------------------------------")
		fmt.Fprintln(w, "Name\tPID\tIP\tCreated\tCommand")
		//fmt.Fprintln(w, "-------------------------------------------------------------------------------")

		for _, c := range containers {
			containerData, err := misc.GetContainerDataFromName(c)
			if err != nil {
				fmt.Printf("Error getting container info: %v\n", err)
				continue
			}
			created := time.Unix(containerData.StartTime, 0).Format("02/01/2006 15:04:05")
			fmt.Fprintf(w,
				"%s\t%d\t%s\t%s\t%s\n", 
				containerData.Name, 
				containerData.PID, 
				containerData.ContainerIP, 
				created, 
				containerData.Command,
			)
		}
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
}
