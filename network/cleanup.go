package network

import (
	"ctrz/runtime"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

const (
	IPSUBNET = "10.200.1.0/24"

	IPTABLES          = "iptables"
	PREROUTING        = "PREROUTING"
	OUTPUT            = "OUTPUT"
	INPUT             = "INPUT"
	POSTROUTING       = "POSTROUTING"
	MASQUERADE        = "MASQUERADE"
	DNAT              = "DNAT"
	FORWARD           = "FORWARD"
	ACCEPT            = "ACCEPT"
	DROP              = "DROP"
	NAT               = "nat"
	PROTOCOL          = "-p"
	TABLE             = "-t"
	DELETE            = "-D"
	TCP               = "tcp"
	JUMP              = "-j"
	SOURCE            = "-s"
	MATCH             = "-m"
	CHECK             = "-C"
	DESTINATION       = "-d"
	DESTINATION_PORT  = "--dport"
	SOURCE_PORT       = "--sport"
	CONNECTION_STATES = "--ctstate"
	CONNTRACK         = "conntrack"
	TO_DESTINATION    = "--to-destination"
)

func RemoveContainerByName(name string, forceKill bool) error {
	dir, err := runtime.CtrzStateDir()
	if err != nil {
		return err
	}
	containerData, err := runtime.GetContainerDataFromName(name)
	if err != nil {
		return err
	}
	_, err = os.Open(fmt.Sprintf("/proc/%d/stat", containerData.PID))
	if err == nil {
		fmt.Printf("Killing process %d\n", containerData.PID)
		if forceKill {
			if err := syscall.Kill(containerData.PID, syscall.SIGKILL); err != nil {
				return fmt.Errorf("Error killing container: %v\n", err)
			}
		} else {
			if err := syscall.Kill(containerData.PID, syscall.SIGTERM); err != nil {
				return fmt.Errorf("Error killing container: %v\n", err)
			}
		}
	}
	if err := removeIPTableRules(containerData); err != nil {
		log.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, "containers", name)); err != nil {
		return err
	}
	if err := RemoveContIP(containerData.ContainerIP); err != nil {
		return err
	}
	return nil
}

func removeIPTableRules(container runtime.ContainerMeta) error {
	for i := range container.ContainerPort {
		cmds := [][]string{
			{
				IPTABLES, TABLE, NAT, DELETE, PREROUTING,
				PROTOCOL, TCP, DESTINATION_PORT, strconv.Itoa(container.HostPort[i]),
				JUMP, DNAT,
				TO_DESTINATION, fmt.Sprintf("%s:%d", container.ContainerIP, container.ContainerPort[i]),
			},
			{
				IPTABLES, TABLE, NAT, DELETE, OUTPUT,
				PROTOCOL, TCP, DESTINATION_PORT, strconv.Itoa(container.HostPort[i]),
				JUMP, DNAT,
				TO_DESTINATION, fmt.Sprintf("%s:%d", container.ContainerIP, container.ContainerPort[i]),
			},
			{
				IPTABLES, TABLE, NAT, DELETE, POSTROUTING, SOURCE,
				IPSUBNET, JUMP, MASQUERADE,
			},
			{
				IPTABLES, DELETE, FORWARD,
				PROTOCOL, TCP, DESTINATION, container.ContainerIP,
				DESTINATION_PORT, strconv.Itoa(container.ContainerPort[i]),
				JUMP, ACCEPT,
			},
			{
				IPTABLES, DELETE, FORWARD,
				PROTOCOL, TCP, SOURCE, container.ContainerIP,
				SOURCE_PORT, strconv.Itoa(container.ContainerPort[i]),
				JUMP, ACCEPT,
			},
			{
				IPTABLES, DELETE, FORWARD,
				MATCH, CONNTRACK,
				CONNECTION_STATES, "ESTABLISHED,RELATED",
				JUMP, ACCEPT,
			},
			{
				IPTABLES, DELETE, INPUT, PROTOCOL, TCP,
				DESTINATION_PORT, strconv.Itoa(container.ContainerPort[i]),
				JUMP, ACCEPT,
			},
			{
				IPTABLES, DELETE, FORWARD,
				DESTINATION, container.ContainerIP,
				JUMP, DROP,
			},
		}

		for _, c := range cmds {
			checkCmd := append([]string(nil), c...)
			for i, v := range checkCmd {
				if v == DELETE {
					checkCmd[i] = CHECK
					break
				}
			}
			out, err := exec.Command(checkCmd[0], checkCmd[1:]...).CombinedOutput()
			if err == nil {
				out, err := exec.Command(c[0], c[1:]...).CombinedOutput()
				if err != nil {
					return fmt.Errorf("%v: %s", err, out)
				}
				continue
			}
			exitErr, ok := err.(*exec.ExitError)
			if !ok {
				return err
			}
			if exitErr.ExitCode() == 1 {
				continue
			}
			return fmt.Errorf("%v: %s", err, out)
		}
	}

	return nil
}
