//go:build linux
// +build linux

package cmd

import (
	"ctrz/misc"
	"ctrz/network"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use: "run",
	Short: "Run a new process in an isolated container",
	Run: func(cmd *cobra.Command, args []string) {
		netns, err := cmd.Flags().GetBool("netns")
		if err != nil {
			log.Fatal("error retrieving --netns. Setting value to default (false): ", err)
			netns = false
		}
		detach, err := cmd.Flags().GetBool("detach")
		if err != nil {
			log.Fatal("error retrieving --detach (-d). Setting value to default (false): ", err)
			detach = false
		}
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatal("error retrieving --name. Value not set", err)
		}
		if name == "" {
			name = misc.GenerateRandomContName()
			for ; !misc.CheckContName(name); {
				name = misc.GenerateRandomContName()
			}
			fmt.Printf("Container name: %s\n", name)
		} else {
			if !misc.CheckContName(name) {
				log.Fatalf("Container '%s' already exists. Either chose a different name or remove the existing container", name)
			}
		}
		pm, err := cmd.Flags().GetStringArray("port") 
		if err != nil {
			log.Fatal("error retrieving --port (-p). Value not set", err)
		}

		period, err := cmd.Flags().GetInt("period")
		if err != nil {
			log.Fatal(err)
		}
		runtime, err := cmd.Flags().GetInt("runtime")
		if err != nil {
			log.Fatal(err)
		}
		var maxCpu string
		if period > 0 && runtime > 0 {
			maxCpu = fmt.Sprintf("%d %d", runtime, period)
		} else {
			cpu, err := cmd.Flags().GetInt("cpu")
			if err != nil {
				log.Fatal(err)
			}
			if cpu <= 0 {
				log.Fatal("--cpu flag must have a value bigger than 0")
			}
			quota := cpu * 1000
			maxCpu = fmt.Sprintf("%d 100000", quota)
		}
		if netns {
			containerIP := "10.200.1.2"
			if len(args) < 1 {
				log.Fatal("At least one command must be provided")
				os.Exit(1)
			}
			pid, cmd, err := network.CreateNetNs(args, maxCpu)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
			if err := network.SetupVeth(pid); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
			if err := syscall.Kill(pid, syscall.SIGCONT); err != nil {
				log.Fatal(err)
			}
			if err := network.DenyAllElse(containerIP); err != nil {
				log.Fatal(err)
			}
			var hostPorts []int
			var containerPorts []int
			for _, p := range pm {
				hostPort, containerPort, err := network.ExposePort(p, containerIP) 
				if err != nil {
					log.Fatal(err)
				}
				go func(hp, cp int) {
    			    if err := network.Userland("tcp", containerIP, hp, cp); err != nil {
    			        log.Printf("proxy %d:%d failed: %v", hp, cp, err)
    			    }
    			}(hostPort, containerPort)

				hostPorts = append(hostPorts, hostPort)
				containerPorts = append(containerPorts, containerPort)
			}
			err = misc.AttachNameToPID(pid, name, args, containerIP, containerPorts, hostPorts)
			if err != nil {
				log.Fatal(err)
			}
			if !detach {
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Wait()
			}
			select{}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().Bool("netns", false, "Create a new network namespace")
	runCmd.Flags().BoolP("detach", "d", false, "detach process from terminal")
	runCmd.Flags().Int("cpu", 100, "Limits the CPU usage (in %) of the wrapped process")
	runCmd.Flags().Int("runtime", 0, "Configures the runtime of a process within a cgroup")
	runCmd.Flags().Int("period", 0, "Determines the time window in which to apply the runtime")
	runCmd.Flags().String("name", "", "name of new container")
	runCmd.Flags().StringArrayP("port", "p", []string{}, "Map host port to container port with '<host-port>:<container-port>'")
	runCmd.MarkFlagsRequiredTogether("runtime", "period")
	runCmd.MarkFlagsMutuallyExclusive("cpu", "runtime")
}