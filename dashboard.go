package main

import (
	"ctrz/cgroup"
	"ctrz/network"
	"ctrz/runtime"
	"ctrz/spec"
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Dashboard struct {
	name        string
	pid         int
	cgroupPath  string
	controllers map[string]bool
	container   spec.Container
}

type previous struct {
	cpuUsec uint64
	cpuTime time.Time
	haveCPU bool

	received    uint64
	sent        uint64
	networkTime time.Time
	haveNetwork bool
}

//go:embed dashboard.html
var dash embed.FS

func StartDashboardServer() error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ctrz/dashboard/{name}", ContStats)
	mux.HandleFunc("GET /ctrz/events/{name}", ContEvents)

	server := &http.Server{
		Addr:    "0.0.0.0:9999",
		Handler: mux,
	}

	return server.ListenAndServe()
}

func ContStats(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	_, err := newDashboard(name)
	if err != nil {
		http.Error(w, "Container not found", http.StatusNotFound)
		return
	}

	page, err := dash.ReadFile("dashboard.html")
	if err != nil {
		http.Error(w, "Could not load dashboard", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, err = w.Write(page)
	if err != nil {
		log.Printf("Error writing dashboard: %v", err)
	}
}

func newDashboard(name string) (*Dashboard, error) {
	containerData, err := runtime.GetContainerDataFromName(name)
	if err != nil {
		return nil, err
	}

	cg := cgroup.New()

	path, err := cg.Path(containerData.PID)
	if err != nil {
		return nil, err
	}

	ctrls, err := cg.EnabledControllers()
	if err != nil {
		return nil, err
	}

	return &Dashboard{
		name:        name,
		pid:         containerData.PID,
		cgroupPath:  path,
		controllers: ctrls,
		container:   containerData,
	}, nil
}

func ContEvents(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	dashboard, err := newDashboard(name)
	if err != nil {
		http.Error(w, "Could not initialize container dashboard",
			http.StatusInternalServerError)
		return
	}

	dashboard.Events(w, r)
}

func (d *Dashboard) Events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var prev previous

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		stats := d.collect(&prev)

		data, err := json.Marshal(stats)
		if err != nil {
			return
		}

		// SSE message format.
		if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
			return
		}

		flusher.Flush()

		// Stop when the browser disconnects.
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
		}
	}
}

func (d *Dashboard) collect(p *previous) spec.Stats {
	now := time.Now()

	stats := spec.Stats{
		Name:      d.name,
		Timestamp: now,
		PID:       d.pid,
		Cgroup:    d.cgroupPath,
		Container: d.container,
	}

	// CPU
	if d.controllers["cpu"] {
		cpu, err := cgroup.ReadCPUStat(d.cgroupPath)

		if err != nil {
			stats.Errors = append(stats.Errors, "CPU: "+err.Error())
		} else {
			var usagePercent *float64

			if p.haveCPU {
				elapsed := now.Sub(p.cpuTime).Microseconds()

				if elapsed > 0 && cpu.UsageUsec >= p.cpuUsec {
					percent := float64(cpu.UsageUsec-p.cpuUsec) /
						float64(elapsed) * 100

					usagePercent = &percent
				}
			}

			p.cpuUsec = cpu.UsageUsec
			p.cpuTime = now
			p.haveCPU = true

			stats.CPU = &spec.CPUStats{
				UsagePercent:  usagePercent,
				UsageUsec:     cpu.UsageUsec,
				NrThrottled:   cpu.NrThrottled,
				ThrottledUsec: cpu.ThrottledUsec,
			}
		}
	}

	// Memory
	if d.controllers["memory"] {
		mem, err := cgroup.ReadMemStat(d.cgroupPath)

		if err != nil {
			stats.Errors = append(stats.Errors, "Memory: "+err.Error())
		} else {
			max := "unlimited"

			if mem.Max != 0 && mem.Max <= uint64(1<<63-1) {
				max = convertBytesToFittingUnit(mem.Max)
			}

			stats.Memory = &spec.MemoryStats{
				Current: convertBytesToFittingUnit(mem.Current),
				Max:     max,
			}
		}
	}

	// Network
	sockets, err := network.ResolveSockets(d.pid)

	if err != nil {
		stats.Errors = append(stats.Errors, "Network: "+err.Error())
	} else {
		var received, sent uint64
		var receivedRate, sentRate uint64

		if len(sockets) > 0 {
			received, err = strconv.ParseUint(
				sockets[0].ReceivedBytes, 10, 64,
			)
			if err != nil {
				stats.Errors = append(stats.Errors, "Invalid received byte counter")
			}

			sent, err = strconv.ParseUint(
				sockets[0].SentBytes, 10, 64,
			)
			if err != nil {
				stats.Errors = append(stats.Errors, "Invalid sent byte counter")
			}

			if p.haveNetwork {
				elapsed := now.Sub(p.networkTime)

				if elapsed > 0 {
					if received >= p.received {
						receivedRate = uint64(
							float64(received-p.received) / elapsed.Seconds(),
						)
					}

					if sent >= p.sent {
						sentRate = uint64(
							float64(sent-p.sent) / elapsed.Seconds(),
						)
					}
				}
			}

			if received >= p.received || !p.haveNetwork {
				p.received = received
			}

			if sent >= p.sent || !p.haveNetwork {
				p.sent = sent
			}

			p.networkTime = now
			p.haveNetwork = true
		}

		connections := make([]spec.Connection, 0)

		for i, s := range sockets {
			if i >= 6 {
				break
			}

			connections = append(connections, spec.Connection{
				State:      s.State,
				Proto:      s.Proto,
				LocalAddr:  s.LocalAddr,
				RemoteAddr: s.RemoteAddr,
				Inode:      s.Inode,
			})
		}

		stats.Network = &spec.NetworkStats{
			TotalConnections:  len(sockets),
			ReceivedPerSecond: convertBytesToFittingUnit(receivedRate) + "/s",
			ReceivedTotal:     convertBytesToFittingUnit(received),
			SentPerSecond:     convertBytesToFittingUnit(sentRate) + "/s",
			SentTotal:         convertBytesToFittingUnit(sent),
			Connections:       connections,
		}
	}

	return stats
}

func convertBytesToFittingUnit(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
