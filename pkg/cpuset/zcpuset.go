package cpuset

// Tasks returns the list of process IDs (PIDs) of the processes in the cpuset.
//
// Corresponds to the "tasks" file in the cpuset directory.
func (c *CPUSet) Tasks() ([]int, error) {
	return intsfile(c.path("tasks"))
}

// AddTasks writes to the "tasks" file of the cpuset.
//
// See Tasks() for the meaning of this field.
func (c *CPUSet) AddTasks(tasks []int) error {
	return writeintsfile(c.path("tasks"), tasks)
}

// NotifyOnRelease reports whether the notify_on_release flag is set for this
// cpuset. If true, that cpuset will receive special handling after it is
// released, that is, after all processes cease using it (i.e., terminate or are
// moved to a different cpuset) and all child cpuset directories have been
// removed.
//
// Corresponds to the "notify_on_release" file in the cpuset directory.
func (c *CPUSet) NotifyOnRelease() (bool, error) {
	return flagfile(c.path("notify_on_release"))
}

// SetNotifyOnRelease writes to the "notify_on_release" file of the cpuset.
//
// See NotifyOnRelease() for the meaning of this field.
func (c *CPUSet) SetNotifyOnRelease(enabled bool) error {
	return writeflagfile(c.path("notify_on_release"), enabled)
}

// EnableNotifyOnRelease sets the "notify_on_release" file to true.
//
// See NotifyOnRelease() for the meaning of this field.
func (c *CPUSet) EnableNotifyOnRelease() error { return c.SetNotifyOnRelease(true) }

// DisableNotifyOnRelease sets the "notify_on_release" file to false.
//
// See NotifyOnRelease() for the meaning of this field.
func (c *CPUSet) DisableNotifyOnRelease() error { return c.SetNotifyOnRelease(false) }

// CPUs returns the set of physical numbers of the CPUs on which processes in
// the cpuset are allowed to execute.
//
// Corresponds to the "cpuset.cpus" file in the cpuset directory.
func (c *CPUSet) CPUs() (Set, error) {
	return listfile(c.path("cpuset.cpus"))
}

// SetCPUs writes to the "cpuset.cpus" file of the cpuset.
//
// See CPUs() for the meaning of this field.
func (c *CPUSet) SetCPUs(s Set) error {
	return writelistfile(c.path("cpuset.cpus"), s)
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
//
// Corresponds to the "cpuset.cpu_exclusive" file in the cpuset directory.
func (c *CPUSet) CPUExclusive() (bool, error) {
	return flagfile(c.path("cpuset.cpu_exclusive"))
}

// SetCPUExclusive writes to the "cpuset.cpu_exclusive" file of the cpuset.
//
// See CPUExclusive() for the meaning of this field.
func (c *CPUSet) SetCPUExclusive(enabled bool) error {
	return writeflagfile(c.path("cpuset.cpu_exclusive"), enabled)
}

// EnableCPUExclusive sets the "cpuset.cpu_exclusive" file to true.
//
// See CPUExclusive() for the meaning of this field.
func (c *CPUSet) EnableCPUExclusive() error { return c.SetCPUExclusive(true) }

// DisableCPUExclusive sets the "cpuset.cpu_exclusive" file to false.
//
// See CPUExclusive() for the meaning of this field.
func (c *CPUSet) DisableCPUExclusive() error { return c.SetCPUExclusive(false) }

// Mems returns the list of memory nodes on which processes in this cpuset are
// allowed to allocate memory.
//
// Corresponds to the "cpuset.mems" file in the cpuset directory.
func (c *CPUSet) Mems() (Set, error) {
	return listfile(c.path("cpuset.mems"))
}

// SetMems writes to the "cpuset.mems" file of the cpuset.
//
// See Mems() for the meaning of this field.
func (c *CPUSet) SetMems(s Set) error {
	return writelistfile(c.path("cpuset.mems"), s)
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
//
// Corresponds to the "cpuset.mem_exclusive" file in the cpuset directory.
func (c *CPUSet) MemExclusive() (bool, error) {
	return flagfile(c.path("cpuset.mem_exclusive"))
}

// SetMemExclusive writes to the "cpuset.mem_exclusive" file of the cpuset.
//
// See MemExclusive() for the meaning of this field.
func (c *CPUSet) SetMemExclusive(enabled bool) error {
	return writeflagfile(c.path("cpuset.mem_exclusive"), enabled)
}

// EnableMemExclusive sets the "cpuset.mem_exclusive" file to true.
//
// See MemExclusive() for the meaning of this field.
func (c *CPUSet) EnableMemExclusive() error { return c.SetMemExclusive(true) }

// DisableMemExclusive sets the "cpuset.mem_exclusive" file to false.
//
// See MemExclusive() for the meaning of this field.
func (c *CPUSet) DisableMemExclusive() error { return c.SetMemExclusive(false) }

// MemHardwall reports whether the cpuset is a Hardwall cpuset (see below).
// Unlike mem_exclusive, there is no constraint on whether cpusets marked
// mem_hardwall may have overlapping memory nodes with sibling or cousin
// cpusets. By default, this is off. Newly created cpusets also initially
// default this to off.
//
// Corresponds to the "cpuset.mem_hardwall" file in the cpuset directory.
func (c *CPUSet) MemHardwall() (bool, error) {
	return flagfile(c.path("cpuset.mem_hardwall"))
}

// SetMemHardwall writes to the "cpuset.mem_hardwall" file of the cpuset.
//
// See MemHardwall() for the meaning of this field.
func (c *CPUSet) SetMemHardwall(enabled bool) error {
	return writeflagfile(c.path("cpuset.mem_hardwall"), enabled)
}

// EnableMemHardwall sets the "cpuset.mem_hardwall" file to true.
//
// See MemHardwall() for the meaning of this field.
func (c *CPUSet) EnableMemHardwall() error { return c.SetMemHardwall(true) }

// DisableMemHardwall sets the "cpuset.mem_hardwall" file to false.
//
// See MemHardwall() for the meaning of this field.
func (c *CPUSet) DisableMemHardwall() error { return c.SetMemHardwall(false) }

// MemoryMigrate reports whether memory migration is enabled.
//
// Corresponds to the "cpuset.memory_migrate" file in the cpuset directory.
func (c *CPUSet) MemoryMigrate() (bool, error) {
	return flagfile(c.path("cpuset.memory_migrate"))
}

// SetMemoryMigrate writes to the "cpuset.memory_migrate" file of the cpuset.
//
// See MemoryMigrate() for the meaning of this field.
func (c *CPUSet) SetMemoryMigrate(enabled bool) error {
	return writeflagfile(c.path("cpuset.memory_migrate"), enabled)
}

// EnableMemoryMigrate sets the "cpuset.memory_migrate" file to true.
//
// See MemoryMigrate() for the meaning of this field.
func (c *CPUSet) EnableMemoryMigrate() error { return c.SetMemoryMigrate(true) }

// DisableMemoryMigrate sets the "cpuset.memory_migrate" file to false.
//
// See MemoryMigrate() for the meaning of this field.
func (c *CPUSet) DisableMemoryMigrate() error { return c.SetMemoryMigrate(false) }

// MemoryPressure reports a measure of how much memory pressure the processes in
// this cpuset are causing. If MemoryPressureEnabled() is false this will always
// be 0.
//
// Corresponds to the "cpuset.memory_pressure" file in the cpuset directory.
func (c *CPUSet) MemoryPressure() (int, error) {
	return intfile(c.path("cpuset.memory_pressure"))
}

// MemoryPressureEnabled reports whether memory pressure calculations are
// enabled for all cpusets in the system. This method only works for the root
// cpuset. By default, this is off.
//
// Corresponds to the "cpuset.memory_pressure_enabled" file in the cpuset directory.
func (c *CPUSet) MemoryPressureEnabled() (bool, error) {
	return flagfile(c.path("cpuset.memory_pressure_enabled"))
}

// SetMemoryPressureEnabled writes to the "cpuset.memory_pressure_enabled" file of the cpuset.
//
// See MemoryPressureEnabled() for the meaning of this field.
func (c *CPUSet) SetMemoryPressureEnabled(enabled bool) error {
	return writeflagfile(c.path("cpuset.memory_pressure_enabled"), enabled)
}

// EnableMemoryPressureEnabled sets the "cpuset.memory_pressure_enabled" file to true.
//
// See MemoryPressureEnabled() for the meaning of this field.
func (c *CPUSet) EnableMemoryPressureEnabled() error { return c.SetMemoryPressureEnabled(true) }

// DisableMemoryPressureEnabled sets the "cpuset.memory_pressure_enabled" file to false.
//
// See MemoryPressureEnabled() for the meaning of this field.
func (c *CPUSet) DisableMemoryPressureEnabled() error { return c.SetMemoryPressureEnabled(false) }

// MemorySpreadPage reports whether pages in the kernel page cache
// (filesystem buffers) are uniformly spread across the cpuset.
// By default, this is off (0) in the top cpuset, and inherited
// from the parent cpuset in newly created cpusets.
//
// Corresponds to the "cpuset.memory_spread_page" file in the cpuset directory.
func (c *CPUSet) MemorySpreadPage() (bool, error) {
	return flagfile(c.path("cpuset.memory_spread_page"))
}

// SetMemorySpreadPage writes to the "cpuset.memory_spread_page" file of the cpuset.
//
// See MemorySpreadPage() for the meaning of this field.
func (c *CPUSet) SetMemorySpreadPage(enabled bool) error {
	return writeflagfile(c.path("cpuset.memory_spread_page"), enabled)
}

// EnableMemorySpreadPage sets the "cpuset.memory_spread_page" file to true.
//
// See MemorySpreadPage() for the meaning of this field.
func (c *CPUSet) EnableMemorySpreadPage() error { return c.SetMemorySpreadPage(true) }

// DisableMemorySpreadPage sets the "cpuset.memory_spread_page" file to false.
//
// See MemorySpreadPage() for the meaning of this field.
func (c *CPUSet) DisableMemorySpreadPage() error { return c.SetMemorySpreadPage(false) }

// MemorySpreadSlab reports whether the kernel slab caches for file I/O
// (directory and inode structures) are uniformly spread across the cpuset. By
// default, this is off (0) in the top cpuset, and inherited from the parent
// cpuset in newly created cpusets.
//
// Corresponds to the "cpuset.memory_spread_slab" file in the cpuset directory.
func (c *CPUSet) MemorySpreadSlab() (bool, error) {
	return flagfile(c.path("cpuset.memory_spread_slab"))
}

// SetMemorySpreadSlab writes to the "cpuset.memory_spread_slab" file of the cpuset.
//
// See MemorySpreadSlab() for the meaning of this field.
func (c *CPUSet) SetMemorySpreadSlab(enabled bool) error {
	return writeflagfile(c.path("cpuset.memory_spread_slab"), enabled)
}

// EnableMemorySpreadSlab sets the "cpuset.memory_spread_slab" file to true.
//
// See MemorySpreadSlab() for the meaning of this field.
func (c *CPUSet) EnableMemorySpreadSlab() error { return c.SetMemorySpreadSlab(true) }

// DisableMemorySpreadSlab sets the "cpuset.memory_spread_slab" file to false.
//
// See MemorySpreadSlab() for the meaning of this field.
func (c *CPUSet) DisableMemorySpreadSlab() error { return c.SetMemorySpreadSlab(false) }

// SchedLoadBalance reports wether the kernel will
// automatically load balance processes in that cpuset over the
// allowed CPUs in that cpuset.  If false the kernel will
// avoid load balancing processes in this cpuset, unless some
// other cpuset with overlapping CPUs has its sched_load_balance
// flag set.
//
// Corresponds to the "cpuset.sched_load_balance" file in the cpuset directory.
func (c *CPUSet) SchedLoadBalance() (bool, error) {
	return flagfile(c.path("cpuset.sched_load_balance"))
}

// SetSchedLoadBalance writes to the "cpuset.sched_load_balance" file of the cpuset.
//
// See SchedLoadBalance() for the meaning of this field.
func (c *CPUSet) SetSchedLoadBalance(enabled bool) error {
	return writeflagfile(c.path("cpuset.sched_load_balance"), enabled)
}

// EnableSchedLoadBalance sets the "cpuset.sched_load_balance" file to true.
//
// See SchedLoadBalance() for the meaning of this field.
func (c *CPUSet) EnableSchedLoadBalance() error { return c.SetSchedLoadBalance(true) }

// DisableSchedLoadBalance sets the "cpuset.sched_load_balance" file to false.
//
// See SchedLoadBalance() for the meaning of this field.
func (c *CPUSet) DisableSchedLoadBalance() error { return c.SetSchedLoadBalance(false) }

// SchedRelaxDomainLevel controls the width of the range of CPUs over which the
// kernel scheduler performs immediate rebalancing of runnable tasks across
// CPUs. If sched_load_balance is disabled, then the setting of
// sched_relax_domain_level does not matter, as no such load balancing is done.
// If sched_load_balance is enabled, then the higher the value of the
// sched_relax_domain_level, the wider the range of CPUs over which immediate
// load balancing is attempted.
//
// Corresponds to the "cpuset.sched_relax_domain_level" file in the cpuset directory.
func (c *CPUSet) SchedRelaxDomainLevel() (int, error) {
	return intfile(c.path("cpuset.sched_relax_domain_level"))
}

// SetSchedRelaxDomainLevel writes to the "cpuset.sched_relax_domain_level" file of the cpuset.
//
// See SchedRelaxDomainLevel() for the meaning of this field.
func (c *CPUSet) SetSchedRelaxDomainLevel(level int) error {
	return writeintfile(c.path("cpuset.sched_relax_domain_level"), level)
}
