package spec

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
