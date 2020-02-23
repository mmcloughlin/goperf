package proc

import (
	"strconv"

	"golang.org/x/sys/unix"
)

// SetPriority sets the priority of process with the given pid, where 0 is the calling process.
func SetPriority(pid, priority int) error {
	_, _, errno := unix.Syscall(unix.SYS_SETPRIORITY, unix.PRIO_PROCESS, uintptr(pid), uintptr(priority))
	if errno != 0 {
		return errno
	}
	return nil
}

// Policy is a linux scheduling policy.
type Policy int

// TODO(mbm): move Policy to linux specific file

// Linux scheduling policies.
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
// Reference: https://github.com/torvalds/linux/blob/c309b6f24222246c18a8b65d3950e6e755440865/Documentation/scheduler/sched-design-CFS.rst#L124-L125
//
//	  - SCHED_NORMAL (traditionally called SCHED_OTHER): The scheduling
//	    policy that is used for regular tasks.
//
const (
	SCHED_OTHER    Policy = 0
	SCHED_FIFO     Policy = 1
	SCHED_RR       Policy = 2
	SCHED_BATCH    Policy = 3
	SCHED_IDLE     Policy = 5
	SCHED_DEADLINE Policy = 6
)

// String represents the policy as one of the SCHED_* constants, if possible.
func (p Policy) String() string {
	switch p {
	case 0:
		return "SCHED_OTHER"
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
