package proc

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// SchedParam is the linux sched_param struct.
//
// Reference: https://github.com/torvalds/linux/blob/0a115e5f23b948be369faf14d3bccab283830f56/include/uapi/linux/sched/types.h#L7-L9
//
//	struct sched_param {
//		int sched_priority;
//	};
//
type SchedParam struct {
	Priority int
}

// SetScheduler sets the linux scheduling policy for process pid, where 0 is the calling process.
func SetScheduler(pid int, policy Policy, param *SchedParam) syscall.Errno {
	_, _, err := unix.Syscall(
		unix.SYS_SCHED_SETSCHEDULER,
		uintptr(pid),
		uintptr(policy),
		uintptr(unsafe.Pointer(param)),
	)
	return err
}
