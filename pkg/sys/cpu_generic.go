// +build !linux

package sys

import (
	"fmt"
	"strconv"

	cpuutil "github.com/shirou/gopsutil/cpu"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

func cpu() (cfg.Configuration, error) {
	procs, err := cpuutil.Info()
	if err != nil {
		return nil, err
	}

	c := cfg.Configuration{}
	for _, proc := range procs {
		c = append(c, processor(proc, int(proc.CPU)))
	}
	return c, nil
}

func processor(proc cpuutil.InfoStat, idx int, perftags ...cfg.Tag) cfg.Entry {
	return cfg.Section(
		cfg.Key("cpu"+strconv.Itoa(idx)),
		fmt.Sprintf("processor %d information", proc.CPU),
		cfg.Property("processor", "processor index", cfg.IntValue(proc.CPU)),
		cfg.Property("vendorid", "vendor id", cfg.StringValue(proc.VendorID), perftags...),
		cfg.Property(
			"family",
			`identifies the type of processor in the system (for intel place the number in front of "86")`,
			cfg.StringValue(proc.Family),
		),
		cfg.Property("model", "model number", cfg.StringValue(proc.Model)),
		cfg.Property("stepping", "version number", cfg.IntValue(proc.Stepping)),
		cfg.Property("modelname", "common name of the processor", cfg.StringValue(proc.ModelName), perftags...),
		cfg.Property("physicalid", "physical processor number", cfg.StringValue(proc.PhysicalID)),
		cfg.Property("coreid", "physical core number within the processor", cfg.StringValue(proc.CoreID)),
		cfg.Property("cores", "number of physical cores", cfg.IntValue(proc.Cores)),
		cfg.Property("frequency", "nominal frequency", cfg.FrequencyValue(proc.Mhz*1e6), perftags...),
		cfg.Property("cachesize", "cache size (level 2)", cfg.BytesValue(proc.CacheSize*1024), perftags...),
		cfg.Property("flags", "processor properties and feature sets", cfg.StringsValue(proc.Flags), perftags...),
	)
}
