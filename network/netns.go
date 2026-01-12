package network

import (
	"ctrz/cgroup"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func CreateNetNs(command []string, maxCpu string) (int, *exec.Cmd, error) {
	args := append([]string{"__ctrz_init"}, command...)

	cmd := exec.Command("/proc/self/exe", args...)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNET,
	}
	//cmd.Env = append(os.Environ(), "CTRZ_NET_READY=0")

	err := cmd.Start()
	if err != nil {
		return 0, nil, err
	}

	pid := cmd.Process.Pid
	fmt.Printf("Started process with PID %d\n", pid)

	err = cgroup.CreateAndAttach(pid, maxCpu)
	if err != nil {
		return pid, nil, err
	}
	
	return pid, cmd, nil
}

func SetupNetns() error {
	if out, err := exec.Command("ip", "link", "set", "lo", "up").CombinedOutput(); err != nil {
		return fmt.Errorf("ip link set lo up failed: %v: %s", err, out)
	}
	if out, err := exec.Command("ip", "addr", "add", "10.200.1.2/24", "dev", "ctrz0").CombinedOutput(); err != nil {
		return fmt.Errorf("ip addr add 10.200.1.2/24 dev ctrz0 failed: %v: %s", err, out)
	}
	if out, err := exec.Command("ip", "link", "set", "ctrz0", "up").CombinedOutput(); err != nil {
		return fmt.Errorf("ip link set ctrz0 up failed: %v: %s", err, out)
	}
	if out, err := exec.Command("ip", "route", "add", "default", "via", "10.200.1.1").CombinedOutput(); err != nil {
		return fmt.Errorf("ip route add default via 10.200.1.1 failed: %v: %s", err, out)
	}
	return nil
}

func ExposePort(ports string, containerIP string) (int, int, error) {
	pm, err := parsePorts(ports)
	if err != nil {
		return -1, -1, err
	}
	log.Printf("Exposing port: %d:%d\n", pm.HostPort, pm.ContainerPort)
	cmds := [][]string{
		{
			"iptables", "-t", "nat", "-I", "PREROUTING", "1",
			"-p", "tcp", "--dport", strconv.Itoa(pm.HostPort),
			"-j", "DNAT",
			"--to-destination", fmt.Sprintf("%s:%d", containerIP, pm.ContainerPort),
		},
		{
			"iptables", "-t", "nat", "-I", "OUTPUT", "1",
			"-p", "tcp", "--dport", strconv.Itoa(pm.HostPort),
			"-j", "DNAT",
			"--to-destination", fmt.Sprintf("%s:%d", containerIP, pm.ContainerPort),
		},
		{
			"iptables", "-t", "nat", "-A", "POSTROUTING", "-s",
			"10.200.1.0/24", "-j", "MASQUERADE",
		},
		{
			"iptables", "-I", "FORWARD", "1",
			"-p", "tcp", "-d", containerIP,
			"--dport", strconv.Itoa(pm.ContainerPort),
			"-j", "ACCEPT",
		},
		{
			"iptables", "-I", "FORWARD", "1",
			"-p", "tcp", "-s", containerIP,
			"--sport", strconv.Itoa(pm.ContainerPort),
			"-j", "ACCEPT",
		},
		{
			"iptables", "-I", "FORWARD", "1",
			"-m", "conntrack",
			"--ctstate", "ESTABLISHED,RELATED",
			"-j", "ACCEPT",
		},
		{
			"iptables", "-I", "INPUT", "1", "-p", "tcp",
			"--dport", strconv.Itoa(pm.ContainerPort),
			"-j", "ACCEPT",
		},
	}

	for _, c := range cmds {
		fmt.Printf("Executing: %v\n", c)
		out, err := exec.Command(c[0], c[1:]...).CombinedOutput()
		fmt.Printf("Command output: %s\n", string(out)) 
		if err != nil {
			return -1, -1, err
		}
	}
	return pm.HostPort, pm.ContainerPort, nil
}

func parsePorts(ports string) (PortMapping, error) {
	parts := strings.Split(ports, ":")
	if len(parts) != 2 {
		return PortMapping{}, fmt.Errorf("invalid port mapping")
	}
	hp, err := strconv.Atoi(parts[0])
	if err != nil {
		return PortMapping{}, fmt.Errorf("Invalid host port: %s: %v", parts[0], err)
	}
	cp, err := strconv.Atoi(parts[1])
	if err != nil {
		return PortMapping{}, fmt.Errorf("Invalid container port: %s: %v", parts[1], err)
	}
	return PortMapping{hp, cp}, nil
}

func DenyAllElse(containerIP string) error {
	cmds := [][]string{
		{
			"iptables", "-I", "FORWARD", "1",
			"-d", containerIP,
			"-j", "DROP",
		},
	}

	for _, c := range cmds {
		out, err := exec.Command(c[0], c[1:]...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%v: %s", err, out)
		}
	}

	return nil
}

