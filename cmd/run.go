//go:build linux
// +build linux

package cmd

import (
	"fmt"
	"log"
	"syscall"

	"ctrz/network"
	"ctrz/runtime"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a new process in an isolated container",
	Run: func(cmd *cobra.Command, args []string) {
		detach, err := cmd.Flags().GetBool("detach")
		if err != nil {
			log.Fatal("error retrieving --detach (-d):", err)
		}
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatal("error retrieving --name:", err)
		}
		remove, err := cmd.Flags().GetBool("rm")
		if err != nil {
			log.Fatal("error retrieving -rm:", err)
		}
		if name == "" {
			name = runtime.GenerateRandomContName()
			for !runtime.CheckContName(name) {
				name = runtime.GenerateRandomContName()
			}
		} else {
			if !runtime.CheckContName(name) {
				log.Fatalf("Container '%s' already exists. Either choose a different name or remove the existing container", name)
			}
		}
		pm, err := cmd.Flags().GetStringArray("port")
		if err != nil {
			log.Fatal("error retrieving --port (-p)", err)
		}

		period, err := cmd.Flags().GetInt("period")
		if err != nil {
			log.Fatal(err)
		}
		rt, err := cmd.Flags().GetInt("runtime")
		if err != nil {
			log.Fatal(err)
		}
		var maxCpu string
		if period > 0 && rt > 0 {
			maxCpu = fmt.Sprintf("%d %d", rt, period)
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
		containerIP, err := network.AssignContIP()
		if err != nil {
			log.Fatal(err)
		}
		if len(args) < 1 {
			log.Fatal("At least one command must be provided")
		}
		if err := network.SetupHostNetworking(); err != nil {
			log.Fatal(err)
		}
		pid, proc, err := network.CreateNetNs(args, maxCpu, name, containerIP, detach)
		if err != nil {
			log.Fatal(err)
		}
		if err := network.SetupVeth(pid); err != nil {
			log.Fatal(err)
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

			hostPorts = append(hostPorts, hostPort)
			containerPorts = append(containerPorts, containerPort)
		}
		err = runtime.AttachNameToPID(pid, name, args, containerIP, containerPorts, hostPorts)
		if err != nil {
			log.Fatal(err)
		}
		if !detach {
			if err := proc.Wait(); err != nil {
				log.Fatal(err)
			}
		}
		if remove && !detach {
			if err := network.RemoveContainerByName(name, false); err != nil {
				log.Fatalf("Error cleaning up container: %v", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolP("detach", "d", false, "detach process from terminal")
	runCmd.Flags().Int("cpu", 100, "Limits the CPU usage (in %) of the wrapped process")
	runCmd.Flags().Int("runtime", 0, "Configures the runtime of a process within a cgroup")
	runCmd.Flags().Int("period", 0, "Determines the time window in which to apply the runtime")
	runCmd.Flags().String("name", "", "name of new container")
	runCmd.Flags().StringArrayP("port", "p", []string{}, "Map host port to container port with '<host-port>:<container-port>'")
	runCmd.Flags().Bool("rm", false, "Container automatically cleans up after finishing")
	runCmd.MarkFlagsRequiredTogether("runtime", "period")
	runCmd.MarkFlagsMutuallyExclusive("cpu", "runtime")
}
