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

// Tasks returns the list of process IDs (PIDs) of the processes in the cpuset.
func (s *CPUSet) Tasks() ([]int, error) {
	return intsfile(s.path("tasks"))
}

// NotifyOnRelease reports whether the notify_on_release flag is set for this
// cpuset. If true, that cpuset will receive special handling after it is
// released, that is, after all processes cease using it (i.e., terminate or are
// moved to a different cpuset) and all child cpuset directories have been
// removed.
func (s *CPUSet) NotifyOnRelease() (bool, error) {
	return flagfile(s.path("notify_on_release"))
}

// CPUs returns the set of physical numbers of the CPUs on which processes in
// the cpuset are allowed to execute.
func (s *CPUSet) CPUs() (Set, error) {
	return listfile(s.path("cpuset.cpus"))
}

// CPUExclusive reports whether the cpuset has exclusive use of its CPUs (no
// sibling or cousin cpuset may overlap CPUs). By default, this is off. Newly
// created cpusets also initially default this to off.
//
// Two cpusets are sibling cpusets if they share the same parent cpuset in the
// hierarchy. Two cpusets are cousin cpusets if neither is the ancestor of the
// other. Regardless of the cpu_exclusive setting, if one cpuset is the ancestor
// of another, and if both of these cpusets have nonempty cpus, then their cpus
// must overlap, because the cpus of any cpuset are always a subset of the cpus
// of its parent cpuset.
func (s *CPUSet) CPUExclusive() (bool, error) {
	return flagfile(s.path("cpuset.cpu_exclusive"))
}

// Mems returns the list of memory nodes on which processes in this cpuset are
// allowed to allocate memory.
func (s *CPUSet) Mems() (Set, error) {
	return listfile(s.path("cpuset.mems"))
}

// MemExclusive reports whether the cpuset has exclusive use of its memory nodes
// (no sibling or cousin may overlap). Also if set, the cpuset is a Hardwall
// cpuset. By default, this is off. Newly created cpusets also initially default
// this to off.
//
// Regardless of the mem_exclusive setting, if one cpuset is the ancestor of
// another, then their memory nodes must overlap, because the memory nodes of
// any cpuset are always a subset of the memory nodes of that cpuset's parent
// cpuset.
func (s *CPUSet) MemExclusive() (bool, error) {
	return flagfile(s.path("cpuset.mem_exclusive"))
}

// MemHardwall reports whether the cpuset is a Hardwall cpuset (see below).
// Unlike mem_exclusive, there is no constraint on whether cpusets marked
// mem_hardwall may have overlapping memory nodes with sibling or cousin
// cpusets. By default, this is off. Newly created cpusets also initially
// default this to off.
func (s *CPUSet) MemHardwall() (bool, error) {
	return flagfile(s.path("cpuset.mem_hardwall"))
}

// MemoryMigrate reports whether memory migration is enabled.
func (s *CPUSet) MemoryMigrate() (bool, error) {
	return flagfile(s.path("cpuset.memory_migrate"))
}

// MemoryPressure reports a measure of how much memory pressure the processes in
// this cpuset are causing. If MemoryPressureEnabled() is false this will always
// be 0.
func (s *CPUSet) MemoryPressure() (int, error) {
	return intfile(s.path("cpuset.memory_pressure"))
}

// MemoryPressureEnabled reports whether memory pressure calculations are
// enabled for all cpusets in the system. This method only works for the root
// cpuset. By default, this is off.
func (s *CPUSet) MemoryPressureEnabled() (bool, error) {
	return flagfile(s.path("cpuset.memory_pressure_enabled"))
}

// MemorySpreadPage reports whether pages in the kernel page cache
// (filesystem buffers) are uniformly spread across the cpuset.
// By default, this is off (0) in the top cpuset, and inherited
// from the parent cpuset in newly created cpusets.
func (s *CPUSet) MemorySpreadPage() (bool, error) {
	return flagfile(s.path("cpuset.memory_spread_page"))
}

// MemorySpreadSlab reports whether the kernel slab caches for file I/O
// (directory and inode structures) are uniformly spread across the cpuset. By
// default, this is off (0) in the top cpuset, and inherited from the parent
// cpuset in newly created cpusets.
func (s *CPUSet) MemorySpreadSlab() (bool, error) {
	return flagfile(s.path("cpuset.memory_spread_slab"))
}

// SchedLoadBalance reports wether the kernel will
// automatically load balance processes in that cpuset over the
// allowed CPUs in that cpuset.  If false the kernel will
// avoid load balancing processes in this cpuset, unless some
// other cpuset with overlapping CPUs has its sched_load_balance
// flag set.
func (s *CPUSet) SchedLoadBalance() (bool, error) {
	return flagfile(s.path("cpuset.sched_load_balance"))
}

// SchedRelaxDomainLevel controls the width of the range of CPUs over which the
// kernel scheduler performs immediate rebalancing of runnable tasks across
// CPUs. If sched_load_balance is disabled, then the setting of
// sched_relax_domain_level does not matter, as no such load balancing is done.
// If sched_load_balance is enabled, then the higher the value of the
// sched_relax_domain_level, the wider the range of CPUs over which immediate
// load balancing is attempted.
func (s *CPUSet) SchedRelaxDomainLevel() (int, error) {
	return intfile(s.path("cpuset.sched_relax_domain_level"))
}

// path returns the full path to name within the cpuset directory.
func (s *CPUSet) path(name string) string {
	return filepath.Join(s.root, name)
}
