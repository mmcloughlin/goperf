package sys

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mmcloughlin/goperf/pkg/cfg"
)

// Reference: https://github.com/torvalds/linux/blob/0f137416247fe92c0779a9ab49e912a7006869e8/Documentation/driver-api/thermal/sysfs-api.rst#L365-L621
//
//	2. sysfs attributes structure
//	=============================
//
//	==	================
//	RO	read only value
//	WO	write only value
//	RW	read/write value
//	==	================
//
//	Thermal sysfs attributes will be represented under /sys/class/thermal.
//	Hwmon sysfs I/F extension is also available under /sys/class/hwmon
//	if hwmon is compiled in or built as a module.
//
//	Thermal zone device sys I/F, created once it's registered::
//
//	  /sys/class/thermal/thermal_zone[0-*]:
//	    |---type:			Type of the thermal zone
//	    |---temp:			Current temperature
//	    |---mode:			Working mode of the thermal zone
//	    |---policy:			Thermal governor used for this zone
//	    |---available_policies:	Available thermal governors for this zone
//	    |---trip_point_[0-*]_temp:	Trip point temperature
//	    |---trip_point_[0-*]_type:	Trip point type
//	    |---trip_point_[0-*]_hyst:	Hysteresis value for this trip point
//	    |---emul_temp:		Emulated temperature set node
//	    |---sustainable_power:      Sustainable dissipatable power
//	    |---k_po:                   Proportional term during temperature overshoot
//	    |---k_pu:                   Proportional term during temperature undershoot
//	    |---k_i:                    PID's integral term in the power allocator gov
//	    |---k_d:                    PID's derivative term in the power allocator
//	    |---integral_cutoff:        Offset above which errors are accumulated
//	    |---slope:                  Slope constant applied as linear extrapolation
//	    |---offset:                 Offset constant applied as linear extrapolation
//
//	Thermal cooling device sys I/F, created once it's registered::
//
//	  /sys/class/thermal/cooling_device[0-*]:
//	    |---type:			Type of the cooling device(processor/fan/...)
//	    |---max_state:		Maximum cooling state of the cooling device
//	    |---cur_state:		Current cooling state of the cooling device
//	    |---stats:			Directory containing cooling device's statistics
//	    |---stats/reset:		Writing any value resets the statistics
//	    |---stats/time_in_state_ms:	Time (msec) spent in various cooling states
//	    |---stats/total_trans:	Total number of times cooling state is changed
//	    |---stats/trans_table:	Cooing state transition table
//
//
//	Then next two dynamic attributes are created/removed in pairs. They represent
//	the relationship between a thermal zone and its associated cooling device.
//	They are created/removed for each successful execution of
//	thermal_zone_bind_cooling_device/thermal_zone_unbind_cooling_device.
//
//	::
//
//	  /sys/class/thermal/thermal_zone[0-*]:
//	    |---cdev[0-*]:		[0-*]th cooling device in current thermal zone
//	    |---cdev[0-*]_trip_point:	Trip point that cdev[0-*] is associated with
//	    |---cdev[0-*]_weight:       Influence of the cooling device in
//					this thermal zone
//
//	Besides the thermal zone device sysfs I/F and cooling device sysfs I/F,
//	the generic thermal driver also creates a hwmon sysfs I/F for each _type_
//	of thermal zone device. E.g. the generic thermal driver registers one hwmon
//	class device and build the associated hwmon sysfs I/F for all the registered
//	ACPI thermal zones.
//
//	::
//
//	  /sys/class/hwmon/hwmon[0-*]:
//	    |---name:			The type of the thermal zone devices
//	    |---temp[1-*]_input:	The current temperature of thermal zone [1-*]
//	    |---temp[1-*]_critical:	The critical trip point of thermal zone [1-*]
//
//	Please read Documentation/hwmon/sysfs-interface.rst for additional information.
//
//	Thermal zone attributes
//	-----------------------
//
//	type
//		Strings which represent the thermal zone type.
//		This is given by thermal zone driver as part of registration.
//		E.g: "acpitz" indicates it's an ACPI thermal device.
//		In order to keep it consistent with hwmon sys attribute; this should
//		be a short, lowercase string, not containing spaces nor dashes.
//		RO, Required
//
//	temp
//		Current temperature as reported by thermal zone (sensor).
//		Unit: millidegree Celsius
//		RO, Required
//
//	mode
//		One of the predefined values in [enabled, disabled].
//		This file gives information about the algorithm that is currently
//		managing the thermal zone. It can be either default kernel based
//		algorithm or user space application.
//
//		enabled
//				  enable Kernel Thermal management.
//		disabled
//				  Preventing kernel thermal zone driver actions upon
//				  trip points so that user application can take full
//				  charge of the thermal management.
//
//		RW, Optional
//
//	policy
//		One of the various thermal governors used for a particular zone.
//
//		RW, Required
//
//	available_policies
//		Available thermal governors which can be used for a particular zone.
//
//		RO, Required
//
//	`trip_point_[0-*]_temp`
//		The temperature above which trip point will be fired.
//
//		Unit: millidegree Celsius
//
//		RO, Optional
//
//	`trip_point_[0-*]_type`
//		Strings which indicate the type of the trip point.
//
//		E.g. it can be one of critical, hot, passive, `active[0-*]` for ACPI
//		thermal zone.
//
//		RO, Optional
//
//	`trip_point_[0-*]_hyst`
//		The hysteresis value for a trip point, represented as an integer
//		Unit: Celsius
//		RW, Optional
//
//	`cdev[0-*]`
//		Sysfs link to the thermal cooling device node where the sys I/F
//		for cooling device throttling control represents.
//
//		RO, Optional
//
//	`cdev[0-*]_trip_point`
//		The trip point in this thermal zone which `cdev[0-*]` is associated
//		with; -1 means the cooling device is not associated with any trip
//		point.
//
//		RO, Optional
//
//	`cdev[0-*]_weight`
//		The influence of `cdev[0-*]` in this thermal zone. This value
//		is relative to the rest of cooling devices in the thermal
//		zone. For example, if a cooling device has a weight double
//		than that of other, it's twice as effective in cooling the
//		thermal zone.
//
//		RW, Optional
//
//	passive
//		Attribute is only present for zones in which the passive cooling
//		policy is not supported by native thermal driver. Default is zero
//		and can be set to a temperature (in millidegrees) to enable a
//		passive trip point for the zone. Activation is done by polling with
//		an interval of 1 second.
//
//		Unit: millidegrees Celsius
//
//		Valid values: 0 (disabled) or greater than 1000
//
//		RW, Optional
//
//	emul_temp
//		Interface to set the emulated temperature method in thermal zone
//		(sensor). After setting this temperature, the thermal zone may pass
//		this temperature to platform emulation function if registered or
//		cache it locally. This is useful in debugging different temperature
//		threshold and its associated cooling action. This is write only node
//		and writing 0 on this node should disable emulation.
//		Unit: millidegree Celsius
//
//		WO, Optional
//
//		  WARNING:
//		    Be careful while enabling this option on production systems,
//		    because userland can easily disable the thermal policy by simply
//		    flooding this sysfs node with low temperature values.
//
//	sustainable_power
//		An estimate of the sustained power that can be dissipated by
//		the thermal zone. Used by the power allocator governor. For
//		more information see Documentation/driver-api/thermal/power_allocator.rst
//
//		Unit: milliwatts
//
//		RW, Optional
//
//	k_po
//		The proportional term of the power allocator governor's PID
//		controller during temperature overshoot. Temperature overshoot
//		is when the current temperature is above the "desired
//		temperature" trip point. For more information see
//		Documentation/driver-api/thermal/power_allocator.rst
//
//		RW, Optional
//
//	k_pu
//		The proportional term of the power allocator governor's PID
//		controller during temperature undershoot. Temperature undershoot
//		is when the current temperature is below the "desired
//		temperature" trip point. For more information see
//		Documentation/driver-api/thermal/power_allocator.rst
//
//		RW, Optional
//
//	k_i
//		The integral term of the power allocator governor's PID
//		controller. This term allows the PID controller to compensate
//		for long term drift. For more information see
//		Documentation/driver-api/thermal/power_allocator.rst
//
//		RW, Optional
//
//	k_d
//		The derivative term of the power allocator governor's PID
//		controller. For more information see
//		Documentation/driver-api/thermal/power_allocator.rst
//
//		RW, Optional
//
//	integral_cutoff
//		Temperature offset from the desired temperature trip point
//		above which the integral term of the power allocator
//		governor's PID controller starts accumulating errors. For
//		example, if integral_cutoff is 0, then the integral term only
//		accumulates error when temperature is above the desired
//		temperature trip point. For more information see
//		Documentation/driver-api/thermal/power_allocator.rst
//
//		Unit: millidegree Celsius
//
//		RW, Optional
//
//	slope
//		The slope constant used in a linear extrapolation model
//		to determine a hotspot temperature based off the sensor's
//		raw readings. It is up to the device driver to determine
//		the usage of these values.
//
//		RW, Optional
//
//	offset
//		The offset constant used in a linear extrapolation model
//		to determine a hotspot temperature based off the sensor's
//		raw readings. It is up to the device driver to determine
//		the usage of these values.
//
//		RW, Optional
//

// Thermal provides system thermal state.
type Thermal struct{}

// Key returns "thermal".
func (Thermal) Key() cfg.Key { return "thermal" }

// Doc for the configuration provider.
func (Thermal) Doc() string { return "thermal zone states" }

// Available checks whether the thermal sysfs files are present.
func (Thermal) Available() bool {
	_, err := os.Stat("/sys/class/thermal/thermal_zone0/type")
	return err == nil
}

// Configuration queries sysfs for thermal zone status.
func (Thermal) Configuration() (cfg.Configuration, error) {
	properties := []fileproperty{
		property("type", parsestring, "concise thermal zone description"),
		property("temp", parsemillicelsius, "current temperature as reported by zone sensor"),
		property("policy", parsestring, "thermal governor used for this zone"),
	}

	dirs, err := filepath.Glob("/sys/class/thermal/thermal_zone*")
	if err != nil {
		return nil, err
	}

	c := cfg.Configuration{}
	for _, dir := range dirs {
		zone := filepath.Base(dir)
		sub, err := parsefiles(dir, properties)
		if err != nil {
			return nil, err
		}
		section := cfg.Section(
			cfg.Key(zone[len("thermal_"):]),
			fmt.Sprintf("thermal status for %s", zone),
			sub...,
		)
		c = append(c, section)
	}
	return c, nil
}

func parsemillicelsius(s string) (cfg.Value, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return cfg.TemperatureValue(float64(n) / 1000), nil
}
