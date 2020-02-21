package sys

import (
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

var IntelPState = cfg.NewProvider("intelpstate", "Intel P-State driver", intelpstate)

func intelpstate() (cfg.Configuration, error) {
	properties := []struct {
		Filename string
		Doc      string
		Parser   func(string) (cfg.Value, error)
	}{
		{"max_perf_pct", "", parseint},
		{"min_perf_pct", "", parseint},
		{"no_turbo", "", parsebool},
		{"num_pstates", "", parseint},
		{"status", "", parsestring},
		{"turbo_pct", "", parseint},
	}
	c := cfg.Configuration{}
	for _, p := range properties {
		filename := "/sys/devices/system/cpu/intel_pstate/" + p.Filename
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
