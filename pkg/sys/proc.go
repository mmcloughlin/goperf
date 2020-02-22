package sys

import (
	"os"
	"strconv"

	"github.com/c9s/goprocinfo/linux"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

const procstatfile = "/proc/self/stat"

// ProcStat is a config provider based on the `/proc/*/stat` file.
type ProcStat struct{}

// Key returns "procstat".
func (ProcStat) Key() cfg.Key { return "procstat" }

// Doc for the configuration provider.
func (ProcStat) Doc() string { return "Linux process status information from /proc/*/stat" }

// Available checks for the /proc/self/stat file.
func (ProcStat) Available() bool {
	_, err := os.Stat(procstatfile)
	return err == nil
}

// Configuration reports performance-critical parameters from the /proc/self/stat file.
func (ProcStat) Configuration() (cfg.Configuration, error) {
	stat, err := linux.ReadProcessStat("/proc/self/stat")
	if err != nil {
		return nil, err
	}

	return cfg.Configuration{
		cfg.Property(
			"priority",
			"nice value from 0 (high) to 39 (low) or for real-time the negated scheduling priority minus one",
			cfg.IntValue(stat.Priority),
		),
		cfg.Property(
			"nice",
			"nice value in the range 19 (low priority) to -20 (high priority)",
			cfg.IntValue(stat.Nice),
		),
		cfg.Property(
			"rtpriority",
			"real-time priority in the range 1 to 99 or 0 for non-real-time",
			cfg.IntValue(stat.RtPriority),
		),
		cfg.Property(
			"policy",
			"",
			policy(stat.Policy),
		),
	}, nil
}

// policy is a linux scheduling policy.
type policy int

// String represents the policy as one of the SCHED_* constants, if possible.
//
// Reference: https://github.com/torvalds/linux/blob/ca7e1fd1026c5af6a533b4b5447e1d2f153e28f2/include/uapi/linux/sched.h#L106-L115
//
//	/*
//	 * Scheduling policies
//	 */
//	#define SCHED_NORMAL		0
//	#define SCHED_FIFO		1
//	#define SCHED_RR		2
//	#define SCHED_BATCH		3
//	/* SCHED_ISO: reserved but not implemented yet */
//	#define SCHED_IDLE		5
//	#define SCHED_DEADLINE		6
//
func (p policy) String() string {
	switch p {
	case 0:
		return "SCHED_NORMAL"
	case 1:
		return "SCHED_FIFO"
	case 2:
		return "SCHED_RR"
	case 3:
		return "SCHED_BATCH"
	case 5:
		return "SCHED_IDLE"
	case 6:
		return "SCHED_DEADLINE"
	}
	return strconv.Itoa(int(p))
}
