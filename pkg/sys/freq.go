package sys

import (
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
	properties := []struct {
		Filename string
		Parser   func(string) (cfg.Value, error)
		Doc      string
	}{
		{"max_perf_pct", parseint, "maximum p-state that will be selected as a percentage of available performance"},
		{"min_perf_pct", parseint, "minimum p-State that will be requested by the driver as a percentage of the max (non-turbo) performance level"},
		{"no_turbo", parsebool, "when true the driver is limited to p-states below the turbo frequency range"},
		{"num_pstates", parseint, "num p-states supported by the hardware"},
		{"status", parsestring, "active/passive/off"},
		{"turbo_pct", parseint, "percentage of the total performance that is supported by hardware that is in the turbo range"},
	}
	c := cfg.Configuration{}
	for _, p := range properties {
		filename := filepath.Join(intelpstateroot, p.Filename)
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
