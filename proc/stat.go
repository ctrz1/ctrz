package proc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// https://man7.org/linux/man-pages/man5/proc_pid_stat.5.html
type ProcStats struct {
	PID                   int    `json:"pid"`                   // 1
	Comm                  string `json:"comm"`                  // 2
	State                 rune   `json:"state"`                 // 3
	PPID                  int    `json:"ppid"`                  // 4
	Pgrp                  int    `json:"pgrp"`                  // 5
	Session               int    `json:"session"`               // 6
	Tty_nr                int    `json:"tty_nr"`                // 7
	Tpgid                 int    `json:"tpgid"`                 // 8
	Flags                 uint   `json:"flags"`                 // 9
	Minflt                uint64 `json:"minflt"`                // 10
	Cminflt               uint64 `json:"cminflt"`               // 11
	Majflt                uint64 `json:"majflt"`                // 12
	Cmajflt               uint64 `json:"cmajflt"`               // 13
	Utime                 uint64 `json:"utime"`                 // 14
	Stime                 uint64 `json:"stime"`                 // 15
	Cutime                int64  `json:"cutime"`                // 16
	Cstime                int64  `json:"cstime"`                // 17
	Priority              int64  `json:"priority"`              // 18
	Nice                  int64  `json:"nice"`                  // 19
	Num_threads           int64  `json:"num_threads"`           // 20
	Itrealvalue           int64  `json:"itrealvalue"`           // 21
	Starttime             uint64 `json:"starttime"`             // 22
	Vsize                 uint64 `json:"vsize"`                 // 23
	Rss                   int64  `json:"rss"`                   // 24
	Rsslim                uint64 `json:"rsslim"`                // 25
	Startcode             uint64 `json:"startcode"`             // 26
	Endcode               uint64 `json:"endcode"`               // 27
	Startstack            uint64 `json:"startstack"`            // 28
	Kstkesp               uint64 `json:"kstkesp"`               // 29
	Kstkeip               uint64 `json:"kstkeip"`               // 30
	Signal                uint64 `json:"signal"`                // 31
	Blocked               uint64 `json:"blocked"`               // 32
	Sigignore             uint64 `json:"sigignore"`             // 33
	Sigcatch              uint64 `json:"sigcatch"`              // 34
	Wchan                 uint64 `json:"wchan"`                 // 35
	Nswap                 uint64 `json:"nswap"`                 // 36
	Cnswap                uint64 `json:"cnswap"`                // 37
	Exit_signal           int    `json:"exit_signal"`           // 38
	Processor             int    `json:"processor"`             // 39
	Rt_priority           uint   `json:"rt_priority"`           // 40
	Policy                uint   `json:"policy"`                // 41
	Delayacct_blkio_ticks uint64 `json:"delayacct_blkio_ticks"` // 42
	Guest_time            uint64 `json:"guest_time"`            // 43
	Cguest_time           int64  `json:"cguest_time"`           // 44
	Start_data            uint64 `json:"start_data"`            // 45
	End_data              uint64 `json:"end_data"`              // 46
	Start_brk             uint64 `json:"start_brk"`             // 47
	Arg_start             uint64 `json:"arg_start"`             // 48
	Arg_end               uint64 `json:"arg_end"`               // 49
	Env_start             uint64 `json:"env_start"`             // 50
	Env_end               uint64 `json:"env_end"`               // 51
	Exit_code             int    `json:"exit_code"`             // 52
}

func ProcessStats(pid int) (ProcStats, error) {
	b, err := os.ReadFile(filepath.Join("/proc", fmt.Sprintf("%d", pid), "stat"))
	if err != nil {
		return ProcStats{}, fmt.Errorf("Error reading stat file: %w", err)
	}

	s := strings.TrimSpace(string(b))

	// pid is before the first '('
	open := strings.IndexByte(s, '(')
	close := strings.LastIndexByte(s, ')')
	if open == -1 || close == -1 || close < open {
		return ProcStats{}, fmt.Errorf("Invalid stat file")
	}

	var p ProcStats

	if _, err := fmt.Sscanf(s[:open], "%d", &p.PID); err != nil {
		return ProcStats{}, err
	}
	p.Comm = s[open+1 : close]

	fields := strings.Fields(s[close+2:])
	if len(fields) != 50 {
		return ProcStats{}, fmt.Errorf("Expected 50 fields after comm, got %d", len(fields))
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
		return ProcStats{}, err
	}

	return p, nil
}
