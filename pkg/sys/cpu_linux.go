package sys

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mmcloughlin/cb/internal/errutil"
	"github.com/mmcloughlin/cb/pkg/cfg"
)

func cpu() (cfg.Configuration, error) {
	return cpucfg("/proc/cpuinfo")
}

func cpucfg(filename string, perftags ...cfg.Tag) (cfg.Configuration, error) {
	procs, err := cpuinfo(filename, perftags...)
	if err != nil {
		return nil, err
	}

	c := cfg.Configuration{}
	for idx, proc := range procs {
		section := cfg.Section(
			cfg.Key("cpu"+strconv.Itoa(idx)),
			fmt.Sprintf("processor %d information", idx),
			proc...,
		)
		c = append(c, section)
	}

	return c, nil
}

func cpuinfo(filename string, perftags ...cfg.Tag) (_ []cfg.Configuration, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer errutil.CheckClose(&err, f)

	configs := []cfg.Configuration{}
	var cur cfg.Configuration

	s := bufio.NewScanner(f)
	for s.Scan() {
		// Parse key: value from line.
		parts := strings.SplitN(s.Text(), ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Did we start a new processor section?
		if key == "processor" {
			if cur != nil {
				configs = append(configs, cur)
			}

			proc, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("parse processor index: %w", err)
			}
			if proc != len(configs) {
				return nil, errutil.AssertionFailure("expected processor index %d", len(configs))
			}

			cur = cfg.Configuration{
				cfg.Property("processor", "processor index", cfg.IntValue(proc)),
			}
			continue
		}

		// Process the key value.
		entry, err := cpuproperty(key, value, perftags...)
		if err != nil {
			return nil, fmt.Errorf("parse %q: %w", key, err)
		}
		if entry != nil {
			cur = append(cur, entry)
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	// Add the last processor.
	configs = append(configs, cur)

	return configs, nil
}

func cpuproperty(key, value string, perftags ...cfg.Tag) (cfg.Entry, error) {
	properties := map[string]fileproperty{
		// amd64
		"vendor_id":   property("vendorid", parsestring, "vendor id", perftags...),
		"cpu family":  property("family", parsestring, `identifies the type of processor in the system (for intel place the number in front of "86")`),
		"model":       property("model", parsestring, "model number"),
		"stepping":    property("stepping", parseint, "version number"),
		"model name":  property("modelname", parsestring, "common name of the processor", perftags...),
		"physical id": property("physicalid", parsestring, "physical processor number"),
		"core id":     property("coreid", parseint, "physical core number within the processor"),
		"cpu cores":   property("cores", parseint, "number of physical cores"),
		"cpu MHz":     property("frequency", parsemhz, "current frequency"),
		"cache size":  property("cachesize", parsecachesize, "cache size (level 2)"),
		"flags":       property("flags", parsestrings, "processor properties and feature sets"),

		// arm64
		"Features":         property("features", parsestrings, "processor properties and feature sets"),
		"CPU implementer":  property("implementer", parsestring, "arm cpuid implementer code"),
		"CPU architecture": property("architecture", parsestring, "arm cpu architecture"),
		"CPU variant":      property("variant", parsestring, "arm cpuid processor revision code"),
		"CPU part":         property("part", parsestring, "arm cpuid part number"),
		"CPU revision":     property("revision", parsestring, "arm cpuid revision or patch number"),
	}

	p, ok := properties[key]
	if !ok {
		return nil, nil
	}

	return p.parse(value)
}

func parsemhz(s string) (cfg.Value, error) {
	mhz, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, err
	}
	return cfg.FrequencyValue(mhz * 1e6), nil
}

func parsecachesize(s string) (cfg.Value, error) {
	if !strings.HasSuffix(s, " KB") {
		return nil, errutil.AssertionFailure("expected cpuinfo cache size in KB unit")
	}
	kb, err := strconv.Atoi(s[:len(s)-3])
	if err != nil {
		return nil, err
	}
	return cfg.BytesValue(kb * 1024), nil
}

func parsestrings(s string) (cfg.Value, error) {
	return cfg.StringsValue(strings.Fields(s)), nil
}
