package proc

import (
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
func SetScheduler(pid int, policy Policy, param *SchedParam) error {
	_, _, errno := unix.Syscall(
		unix.SYS_SCHED_SETSCHEDULER,
		uintptr(pid),
		uintptr(policy),
		uintptr(unsafe.Pointer(param)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// Affinity returns the CPU affinity of the given process.
func Affinity(pid int) ([]int, error) {
	var s unix.CPUSet
	if err := unix.SchedGetaffinity(pid, &s); err != nil {
		return nil, err
	}
	return cpulist(&s), nil
}

// AffinitySelf returns the
func AffinitySelf() ([]int, error) {
	return Affinity(0)
}

// cpulist converts the set s to a list.
func cpulist(s *unix.CPUSet) []int {
	cpus := []int{}
	n := s.Count()
	for cpu := 0; len(cpus) < n; cpu++ {
		if s.IsSet(cpu) {
			cpus = append(cpus, cpu)
		}
	}
	return cpus
}
