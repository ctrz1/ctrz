//go:build linux

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"ctrz/cgroup"
	"ctrz/misc"
	"ctrz/network"

	"golang.org/x/sys/unix"
)

func ctrzDeamon() {
	ptr, err := syscall.BytePtrFromString("ctrzd")
	if err != nil {
		log.Fatal(err)
	}
	if err := unix.Prctl(unix.PR_SET_NAME, uintptr(unsafe.Pointer(
		ptr,
	)), 0, 0, 0); err != nil {
		slog.Error("Couldn't set process name of ctrz daemon")
	}

	var prevRecBytes uint64 = 0
	var prevSentBytes uint64 = 0

	var prevUsec uint64 = 0
	prevTime := time.Now()

	// CPU usage (%), CPU usec, CPU nr thorttled, CPU throttled usec, current memory usage (KB), max memory usage (KB),
	// recieved network traffic (delta), received network traffic (total), sent network traffic (delta), sent network traffic (total)

	for {
		containers, err := misc.RetrieveAllContainers()
		if err != nil || len(containers) == 0 || containers == nil {
			time.Sleep(time.Second)
			continue
		}
		for _, container := range containers {

			var stats []string

			containerData, err := misc.GetContainerDataFromName(container)
			if err != nil {
				continue
			}

			path, err := cgroup.PathForPID(containerData.PID)
			if err != nil {
				log.Fatal(err)
			}
			ctrls, err := cgroup.EnabledControllers(path)
			if err != nil {
				log.Fatal(err)
			}

			if ctrls["cpu"] {
				cpu, err := cgroup.ReadCPUStat(path)
				if err == nil {
					now := time.Now()
					cpuUsagePercent := float64(cpu.UsageUsec-prevUsec) / float64(now.Sub(prevTime).Microseconds()) * 100

					stats = append(stats,
						strconv.FormatFloat(cpuUsagePercent, 'f', 2, 64),
						strconv.FormatUint(cpu.UsageUsec, 10),
						strconv.FormatUint(cpu.NrThrottled, 10),
						strconv.FormatUint(cpu.ThrottledUsec, 10),
					)

					prevUsec = cpu.UsageUsec
					prevTime = now
				}
			}

			if ctrls["memory"] {
				mem, err := cgroup.ReadMemStat(path)
				if err == nil {
					stats = append(stats, strconv.FormatFloat(float64(mem.Current/1024), 'f', 2, 64))
					if mem.Max <= math.MaxInt64-1024 && mem.Max != 0 {
						stats = append(stats, strconv.FormatFloat(float64(mem.Max/1024), 'f', 2, 64))
					}
				}
			}

			sockets, err := network.ResolveSockets(containerData.PID)
			if err == nil && len(sockets) > 0 {
				currentSent, _ := strconv.ParseUint(sockets[0].ReceivedBytes, 10, 64)
				currentReceived, _ := strconv.ParseUint(sockets[0].SentBytes, 10, 64)

				deltaSent := currentSent - prevSentBytes
				deltaReceived := currentReceived - prevRecBytes

				stats = append(stats,
					strconv.Itoa(len(sockets)),
					strconv.FormatUint(deltaReceived, 10),
					strconv.FormatUint(currentReceived, 10),
					strconv.FormatUint(deltaSent, 10),
					strconv.FormatUint(currentSent, 10),
					fmt.Sprintf("%v", sockets),
				)
				prevRecBytes = currentReceived
				prevSentBytes = currentSent
			}
			fmt.Printf("%v\n", stats)
			printStatsToFile(container, stats...)
		}
		time.Sleep(time.Second)
	}
}

func printStatsToFile(containerName string, elems ...string) {
	path := fmt.Sprintf("/var/lib/ctrz/containers/%s/stats.csv", containerName)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o755)
	if err != nil {
		return
	}
	w := csv.NewWriter(f)
	_ = w.Write(elems)
	w.Flush()
}
