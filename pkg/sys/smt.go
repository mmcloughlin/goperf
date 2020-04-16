package sys

import (
	"github.com/mmcloughlin/cb/pkg/cfg"
	"github.com/mmcloughlin/cb/pkg/proc"
	"github.com/mmcloughlin/cb/pkg/pseudofs"
)

// Reference: https://github.com/torvalds/linux/blob/34dabd81160f7bfb18b67c1161b3c4d7ca6cab83/Documentation/ABI/testing/sysfs-devices-system-cpu#L511-L531
//
//	What:		/sys/devices/system/cpu/smt
//			/sys/devices/system/cpu/smt/active
//			/sys/devices/system/cpu/smt/control
//	Date:		June 2018
//	Contact:	Linux kernel mailing list <linux-kernel@vger.kernel.org>
//	Description:	Control Symetric Multi Threading (SMT)
//
//			active:  Tells whether SMT is active (enabled and siblings online)
//
//			control: Read/write interface to control SMT. Possible
//				 values:
//
//				 "on"		  SMT is enabled
//				 "off"		  SMT is disabled
//				 "forceoff"	  SMT is force disabled. Cannot be changed.
//				 "notsupported"   SMT is not supported by the CPU
//				 "notimplemented" SMT runtime toggling is not
//						  implemented for the architecture
//
//				 If control status is "forceoff" or "notsupported" writes
//				 are rejected.
//

const (
	smtroot    = "/sys/devices/system/cpu/smt"
	smtactive  = smtroot + "/active"
	smtcontrol = smtroot + "/control"
)

// SMT provides simultaneous multithreading configuration.
type SMT struct{}

// Key returns "smt".
func (SMT) Key() cfg.Key { return "smt" }

// Doc for the configuration provider.
func (SMT) Doc() string { return "simultaneous multithreading configuration" }

// Available checks whether the smt sysfs files are present.
func (SMT) Available() bool { return proc.Readable(smtactive) }

// Configuration queries sysfs for simultaneous multithreading configuration.
func (SMT) Configuration() (cfg.Configuration, error) {
	return parsefiles(smtroot, []fileproperty{
		perfproperty("active", parsebool, "whether smt is active (enabled and siblings online)"),
	})
}

// DeactivateSMT disables simultaneous multithreading.
type DeactivateSMT struct{}

// Name of the tuning method.
func (DeactivateSMT) Name() string { return "deactivate_smt" }

// Available reports whether the process can control SMT.
func (DeactivateSMT) Available() bool {
	setting, err := pseudofs.String(smtcontrol)
	if err != nil {
		return false
	}
	switch setting {
	case "forceoff", "notsupported", "notimplemented":
		return false
	}
	return proc.Writable(smtcontrol)
}

// Apply turns off SMT.
func (DeactivateSMT) Apply() error {
	return pseudofs.WriteString(smtcontrol, "off")
}

// Reset turns on SMT.
func (DeactivateSMT) Reset() error { return pseudofs.WriteString(smtcontrol, "on") }
