// Package cpuset is a library for manipulation of Linux cpusets.
package cpuset

import (
	"path/filepath"

	"golang.org/x/sys/unix"
)

const stdbase = "/sys/fs/cgroup/cpuset"

// CPUSet represents a cpuset in the sysfs filesystem.
type CPUSet struct {
	root string
}

// Root returns the root cpuset.
func Root() *CPUSet {
	return NewCPUSet("")
}

// NewCPUSet returns a reference to the named cpuset under the standard sysfs hierarchy.
func NewCPUSet(name string) *CPUSet {
	return NewCPUSetPath(filepath.Join(stdbase, name))
}

// NewCPUSetPath returns a reference to a cpuset directory at a custom path.
func NewCPUSetPath(path string) *CPUSet {
	return &CPUSet{
		root: path,
	}
}

// Create the named cpuset under the standard sysfs hierarchy.
func Create(name string) (*CPUSet, error) {
	return CreatePath(filepath.Join(stdbase, name))
}

// CreatePath creates a cpuset at a custom path.
func CreatePath(path string) (*CPUSet, error) {
	if err := unix.Mkdir(path, 0o755); err != nil {
		return nil, err
	}
	return NewCPUSetPath(path), nil
}

// Path returns the path to the cpuset directory.
func (c *CPUSet) Path() string {
	return c.path("")
}

// Remove the cpuset. Note the cpuset must have no children or attached
// processes.
func (c *CPUSet) Remove() error {
	return unix.Rmdir(c.root)
}

// AddTask adds a single task to the cpuset.
func (c *CPUSet) AddTask(task int) error {
	return c.AddTasks([]int{task})
}

// path returns the full path to name within the cpuset directory.
func (c *CPUSet) path(name string) string {
	return filepath.Join(c.root, name)
}

//go:generate go run make_cpuset.go -output zcpuset.go
