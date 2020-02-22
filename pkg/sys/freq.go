package sys

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

// Reference: https://github.com/torvalds/linux/blob/4dd2ab9a0f84a446c65ff33c95339f1cd0e21a4b/Documentation/admin-guide/pm/intel_pstate.rst#L321-L426
//
//	User Space Interface in ``sysfs``
//	=================================
//
//	Global Attributes
//	-----------------
//
//	``intel_pstate`` exposes several global attributes (files) in ``sysfs`` to
//	control its functionality at the system level.  They are located in the
//	``/sys/devices/system/cpu/intel_pstate/`` directory and affect all CPUs.
//
//	Some of them are not present if the ``intel_pstate=per_cpu_perf_limits``
//	argument is passed to the kernel in the command line.
//
//	``max_perf_pct``
//		Maximum P-state the driver is allowed to set in percent of the
//		maximum supported performance level (the highest supported `turbo
//		P-state <turbo_>`_).
//
//		This attribute will not be exposed if the
//		``intel_pstate=per_cpu_perf_limits`` argument is present in the kernel
//		command line.
//
//	``min_perf_pct``
//		Minimum P-state the driver is allowed to set in percent of the
//		maximum supported performance level (the highest supported `turbo
//		P-state <turbo_>`_).
//
//		This attribute will not be exposed if the
//		``intel_pstate=per_cpu_perf_limits`` argument is present in the kernel
//		command line.
//
//	``num_pstates``
//		Number of P-states supported by the processor (between 0 and 255
//		inclusive) including both turbo and non-turbo P-states (see
//		`Turbo P-states Support`_).
//
//		The value of this attribute is not affected by the ``no_turbo``
//		setting described `below <no_turbo_attr_>`_.
//
//		This attribute is read-only.
//
//	``turbo_pct``
//		Ratio of the `turbo range <turbo_>`_ size to the size of the entire
//		range of supported P-states, in percent.
//
//		This attribute is read-only.
//
//	.. _no_turbo_attr:
//
//	``no_turbo``
//		If set (equal to 1), the driver is not allowed to set any turbo P-states
//		(see `Turbo P-states Support`_).  If unset (equalt to 0, which is the
//		default), turbo P-states can be set by the driver.
//		[Note that ``intel_pstate`` does not support the general ``boost``
//		attribute (supported by some other scaling drivers) which is replaced
//		by this one.]
//
//		This attrubute does not affect the maximum supported frequency value
//		supplied to the ``CPUFreq`` core and exposed via the policy interface,
//		but it affects the maximum possible value of per-policy P-state	limits
//		(see `Interpretation of Policy Attributes`_ below for details).
//
//	``hwp_dynamic_boost``
//		This attribute is only present if ``intel_pstate`` works in the
//		`active mode with the HWP feature enabled <Active Mode With HWP_>`_ in
//		the processor.  If set (equal to 1), it causes the minimum P-state limit
//		to be increased dynamically for a short time whenever a task previously
//		waiting on I/O is selected to run on a given logical CPU (the purpose
//		of this mechanism is to improve performance).
//
//		This setting has no effect on logical CPUs whose minimum P-state limit
//		is directly set to the highest non-turbo P-state or above it.
//
//	.. _status_attr:
//
//	``status``
//		Operation mode of the driver: "active", "passive" or "off".
//
//		"active"
//			The driver is functional and in the `active mode
//			<Active Mode_>`_.
//
//		"passive"
//			The driver is functional and in the `passive mode
//			<Passive Mode_>`_.
//
//		"off"
//			The driver is not functional (it is not registered as a scaling
//			driver with the ``CPUFreq`` core).
//
//		This attribute can be written to in order to change the driver's
//		operation mode or to unregister it.  The string written to it must be
//		one of the possible values of it and, if successful, the write will
//		cause the driver to switch over to the operation mode represented by
//		that string - or to be unregistered in the "off" case.  [Actually,
//		switching over from the active mode to the passive mode or the other
//		way around causes the driver to be unregistered and registered again
//		with a different set of callbacks, so all of its settings (the global
//		as well as the per-policy ones) are then reset to their default
//		values, possibly depending on the target operation mode.]
//
//		That only is supported in some configurations, though (for example, if
//		the `HWP feature is enabled in the processor <Active Mode With HWP_>`_,
//		the operation mode of the driver cannot be changed), and if it is not
//		supported in the current configuration, writes to this attribute will
//		fail with an appropriate error.
//

const intelpstateroot = "/sys/devices/system/cpu/intel_pstate"

// IntelPState provides configuration about the Intel P-State driver.
type IntelPState struct{}

// Key returns "intelpstate".
func (IntelPState) Key() cfg.Key { return "intelpstate" }

// Doc for the configuration provider.
func (IntelPState) Doc() string { return "Intel P-State driver" }

// Available checks whether the Intel P-State sysfs files are present.
func (IntelPState) Available() bool {
	info, err := os.Stat(intelpstateroot)
	return err == nil && info.IsDir()
}

// Configuration queries sysfs for Intel P-state configuration.
func (IntelPState) Configuration() (cfg.Configuration, error) {
	return parsefiles(intelpstateroot, []fileproperty{
		{"max_perf_pct", "", parseint, "maximum p-state that will be selected as a percentage of available performance"},
		{"min_perf_pct", "", parseint, "minimum p-State that will be requested by the driver as a percentage of the max (non-turbo) performance level"},
		{"no_turbo", "", parsebool, "when true the driver is limited to p-states below the turbo frequency range"},
		{"num_pstates", "", parseint, "num p-states supported by the hardware"},
		{"status", "", parsestring, "active/passive/off"},
		{"turbo_pct", "", parseint, "percentage of the total performance that is supported by hardware that is in the turbo range"},
	})
}

// Reference: https://github.com/torvalds/linux/blob/4dd2ab9a0f84a446c65ff33c95339f1cd0e21a4b/Documentation/admin-guide/pm/cpufreq.rst#L224-L331
//
//	``affected_cpus``
//		List of online CPUs belonging to this policy (i.e. sharing the hardware
//		performance scaling interface represented by the ``policyX`` policy
//		object).
//
//	``bios_limit``
//		If the platform firmware (BIOS) tells the OS to apply an upper limit to
//		CPU frequencies, that limit will be reported through this attribute (if
//		present).
//
//		The existence of the limit may be a result of some (often unintentional)
//		BIOS settings, restrictions coming from a service processor or another
//		BIOS/HW-based mechanisms.
//
//		This does not cover ACPI thermal limitations which can be discovered
//		through a generic thermal driver.
//
//		This attribute is not present if the scaling driver in use does not
//		support it.
//
//	``cpuinfo_cur_freq``
//		Current frequency of the CPUs belonging to this policy as obtained from
//		the hardware (in KHz).
//
//		This is expected to be the frequency the hardware actually runs at.
//		If that frequency cannot be determined, this attribute should not
//		be present.
//
//	``cpuinfo_max_freq``
//		Maximum possible operating frequency the CPUs belonging to this policy
//		can run at (in kHz).
//
//	``cpuinfo_min_freq``
//		Minimum possible operating frequency the CPUs belonging to this policy
//		can run at (in kHz).
//
//	``cpuinfo_transition_latency``
//		The time it takes to switch the CPUs belonging to this policy from one
//		P-state to another, in nanoseconds.
//
//		If unknown or if known to be so high that the scaling driver does not
//		work with the `ondemand`_ governor, -1 (:c:macro:`CPUFREQ_ETERNAL`)
//		will be returned by reads from this attribute.
//
//	``related_cpus``
//		List of all (online and offline) CPUs belonging to this policy.
//
//	``scaling_available_governors``
//		List of ``CPUFreq`` scaling governors present in the kernel that can
//		be attached to this policy or (if the |intel_pstate| scaling driver is
//		in use) list of scaling algorithms provided by the driver that can be
//		applied to this policy.
//
//		[Note that some governors are modular and it may be necessary to load a
//		kernel module for the governor held by it to become available and be
//		listed by this attribute.]
//
//	``scaling_cur_freq``
//		Current frequency of all of the CPUs belonging to this policy (in kHz).
//
//		In the majority of cases, this is the frequency of the last P-state
//		requested by the scaling driver from the hardware using the scaling
//		interface provided by it, which may or may not reflect the frequency
//		the CPU is actually running at (due to hardware design and other
//		limitations).
//
//		Some architectures (e.g. ``x86``) may attempt to provide information
//		more precisely reflecting the current CPU frequency through this
//		attribute, but that still may not be the exact current CPU frequency as
//		seen by the hardware at the moment.
//
//	``scaling_driver``
//		The scaling driver currently in use.
//
//	``scaling_governor``
//		The scaling governor currently attached to this policy or (if the
//		|intel_pstate| scaling driver is in use) the scaling algorithm
//		provided by the driver that is currently applied to this policy.
//
//		This attribute is read-write and writing to it will cause a new scaling
//		governor to be attached to this policy or a new scaling algorithm
//		provided by the scaling driver to be applied to it (in the
//		|intel_pstate| case), as indicated by the string written to this
//		attribute (which must be one of the names listed by the
//		``scaling_available_governors`` attribute described above).
//
//	``scaling_max_freq``
//		Maximum frequency the CPUs belonging to this policy are allowed to be
//		running at (in kHz).
//
//		This attribute is read-write and writing a string representing an
//		integer to it will cause a new limit to be set (it must not be lower
//		than the value of the ``scaling_min_freq`` attribute).
//
//	``scaling_min_freq``
//		Minimum frequency the CPUs belonging to this policy are allowed to be
//		running at (in kHz).
//
//		This attribute is read-write and writing a string representing a
//		non-negative integer to it will cause a new limit to be set (it must not
//		be higher than the value of the ``scaling_max_freq`` attribute).
//
//	``scaling_setspeed``
//		This attribute is functional only if the `userspace`_ scaling governor
//		is attached to the given policy.
//
//		It returns the last frequency requested by the governor (in kHz) or can
//		be written to in order to set a new frequency for the policy.
//

// CPUFreq provides configuration about CPU frequency scaling.
type CPUFreq struct{}

// Key returns "cpufreq".
func (CPUFreq) Key() cfg.Key { return "cpufreq" }

// Doc for the configuration provider.
func (CPUFreq) Doc() string { return "CPU frequency scaling status" }

// Available checks whether the cpufreq sysfs files are present.
func (CPUFreq) Available() bool {
	_, err := os.Stat("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor")
	return err == nil
}

// Configuration queries sysfs for CPU frequency scaling status.
func (CPUFreq) Configuration() (cfg.Configuration, error) {
	properties := []fileproperty{
		{"cpuinfo_min_freq", "", parsekhz, "minimum operating frequency the processor can run at"},
		{"cpuinfo_max_freq", "", parsekhz, "maximum operating frequency the processor can run at"},
		{"cpuinfo_transition_latency", "", parseint, "time it takes on this cpu to switch between two frequencies in nanoseconds"},
		{"scaling_driver", "", parsestring, "which cpufreq driver is used to set the frequency on this cpu"},
		{"scaling_governor", "", parsestring, "currently active scaling governor on this cpu"},
		{"scaling_min_freq", "", parsekhz, "minimum allowed frequency by the current scaling policy"},
		{"scaling_min_freq", "", parsekhz, "maximum allowed frequency by the current scaling policy"},
		{"scaling_cur_freq", "", parsekhz, "current frequency as determined by the governor and cpufreq core"},
	}

	dirs, err := filepath.Glob("/sys/devices/system/cpu/cpu*/cpufreq")
	if err != nil {
		return nil, err
	}

	c := cfg.Configuration{}
	for _, dir := range dirs {
		cpu := filepath.Base(filepath.Dir(dir))
		sub, err := parsefiles(dir, properties)
		if err != nil {
			return nil, err
		}
		section := cfg.Section(
			cfg.Key(cpu),
			fmt.Sprintf("cpu frequency status for %s", cpu),
			sub...,
		)
		c = append(c, section)
	}
	return c, nil
}
