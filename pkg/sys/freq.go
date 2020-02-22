package sys

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

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
		{"max_perf_pct", parseint, "maximum p-state that will be selected as a percentage of available performance"},
		{"min_perf_pct", parseint, "minimum p-State that will be requested by the driver as a percentage of the max (non-turbo) performance level"},
		{"no_turbo", parsebool, "when true the driver is limited to p-states below the turbo frequency range"},
		{"num_pstates", parseint, "num p-states supported by the hardware"},
		{"status", parsestring, "active/passive/off"},
		{"turbo_pct", parseint, "percentage of the total performance that is supported by hardware that is in the turbo range"},
	})
}

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
		{"cpuinfo_min_freq", parsekhz, "minimum operating frequency the processor can run at"},
		{"cpuinfo_max_freq", parsekhz, "maximum operating frequency the processor can run at"},
		{"cpuinfo_transition_latency", parseint, "time it takes on this cpu to switch between two frequencies in nanoseconds"},
		{"scaling_driver", parsestring, "which cpufreq driver is used to set the frequency on this cpu"},
		{"scaling_governor", parsestring, "currently active scaling governor on this cpu"},
		{"scaling_min_freq", parsekhz, "minimum allowed frequency by the current scaling policy"},
		{"scaling_min_freq", parsekhz, "maximum allowed frequency by the current scaling policy"},
		{"scaling_cur_freq", parsekhz, "current frequency as determined by the governor and cpufreq core"},
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

type fileproperty struct {
	Filename string
	Parser   func(string) (cfg.Value, error)
	Doc      string
}

func parsefiles(root string, properties []fileproperty) (cfg.Configuration, error) {
	c := cfg.Configuration{}
	for _, p := range properties {
		filename := filepath.Join(root, p.Filename)
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		data := strings.TrimSpace(string(b))
		v, err := p.Parser(data)
		if err != nil {
			return nil, err
		}
		c = append(c, cfg.Property(
			cfg.Key(strings.ReplaceAll(p.Filename, "_", "")),
			p.Doc,
			v,
		))
	}
	return c, nil
}

func parseint(s string) (cfg.Value, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return cfg.IntValue(n), nil
}

func parsekhz(s string) (cfg.Value, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return cfg.FrequencyValue(n * 1000), nil
}

func parsebool(s string) (cfg.Value, error) {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return nil, err
	}
	return cfg.BoolValue(b), nil
}

func parsestring(s string) (cfg.Value, error) {
	return cfg.StringValue(s), nil
}
