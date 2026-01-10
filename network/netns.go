package network

import (
	"ctrz/cgroup"
	"fmt"
	"os/exec"
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
