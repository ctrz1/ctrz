package proc

import (
	"ctrz/spec"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ProcessStats(pid int) (spec.ProcStats, error) {
	b, err := os.ReadFile(filepath.Join("/proc", fmt.Sprintf("%d", pid), "stat"))
	if err != nil {
		return spec.ProcStats{}, fmt.Errorf("Error reading stat file: %w", err)
	}

	s := strings.TrimSpace(string(b))

	// pid is before the first '('
	open := strings.IndexByte(s, '(')
	close := strings.LastIndexByte(s, ')')
	if open == -1 || close == -1 || close < open {
		return spec.ProcStats{}, fmt.Errorf("Invalid stat file")
	}

	var p spec.ProcStats

	if _, err := fmt.Sscanf(s[:open], "%d", &p.PID); err != nil {
		return spec.ProcStats{}, err
	}
	p.Comm = s[open+1 : close]

	fields := strings.Fields(s[close+2:])
	if len(fields) != 50 {
		return spec.ProcStats{}, fmt.Errorf("Expected 50 fields after comm, got %d", len(fields))
	}

	_, err = fmt.Sscanf(strings.Join(fields, " "),
		"%c %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d",
		&p.State,                 // 3
		&p.PPID,                  // 4
		&p.Pgrp,                  // 5
		&p.Session,               // 6
		&p.Tty_nr,                // 7
		&p.Tpgid,                 // 8
		&p.Flags,                 // 9
		&p.Minflt,                // 10
		&p.Cminflt,               // 11
		&p.Majflt,                // 12
		&p.Cmajflt,               // 13
		&p.Utime,                 // 14
		&p.Stime,                 // 15
		&p.Cutime,                // 16
		&p.Cstime,                // 17
		&p.Priority,              // 18
		&p.Nice,                  // 19
		&p.Num_threads,           // 20
		&p.Itrealvalue,           // 21
		&p.Starttime,             // 22
		&p.Vsize,                 // 23
		&p.Rss,                   // 24
		&p.Rsslim,                // 25
		&p.Startcode,             // 26
		&p.Endcode,               // 27
		&p.Startstack,            // 28
		&p.Kstkesp,               // 29
		&p.Kstkeip,               // 30
		&p.Signal,                // 31
		&p.Blocked,               // 32
		&p.Sigignore,             // 33
		&p.Sigcatch,              // 34
		&p.Wchan,                 // 35
		&p.Nswap,                 // 36
		&p.Cnswap,                // 37
		&p.Exit_signal,           // 38
		&p.Processor,             // 39
		&p.Rt_priority,           // 40
		&p.Policy,                // 41
		&p.Delayacct_blkio_ticks, // 42
		&p.Guest_time,            // 43
		&p.Cguest_time,           // 44
		&p.Start_data,            // 45
		&p.End_data,              // 46
		&p.Start_brk,             // 47
		&p.Arg_start,             // 48
		&p.Arg_end,               // 49
		&p.Env_start,             // 50
		&p.Env_end,               // 51
		&p.Exit_code,             // 52
	)
	if err != nil {
		return spec.ProcStats{}, err
	}

	return p, nil
}
