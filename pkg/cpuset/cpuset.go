package cpuset

import (
	"path/filepath"
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

// path returns the full path to name within the cpuset directory.
func (s *CPUSet) path(name string) string {
	return filepath.Join(s.root, name)
}

//go:generate go run make_cpuset.go -output zcpuset.go
