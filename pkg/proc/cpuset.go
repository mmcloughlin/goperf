package proc

import (
	"golang.org/x/sys/unix"

	"github.com/mmcloughlin/cb/pkg/cpuset"
)

// SetCPUSet moves pid to the named cpuset.
func SetCPUSet(pid int, name string) error {
	s := cpuset.NewCPUSet(name)
	return s.AddTask(pid)
}

// SetCPUSetSelf moves this thread to the named cpuset.
func SetCPUSetSelf(name string) error {
	return SetCPUSet(unix.Gettid(), name)
}
