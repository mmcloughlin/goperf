package sys

import (
	"github.com/shirou/gopsutil/mem"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

var VirtualMemory = cfg.NewProvider(
	"mem",
	"Virtual memory usage statistics.",
	virtualmemory,
)

func virtualmemory() (cfg.Configuration, error) {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return cfg.Configuration{
		cfg.Property(
			"total",
			"Total amount of RAM on this system.",
			cfg.BytesValue(vmem.Total),
		),
		cfg.Property(
			"available",
			"RAM available for programs to allocate.",
			cfg.BytesValue(vmem.Available),
		),
		cfg.Property(
			"used",
			"RAM used by programs.",
			cfg.BytesValue(vmem.Used),
		),
		cfg.Property(
			"used-percent",
			"Percentage of RAM used by programs.",
			cfg.PercentageValue(vmem.UsedPercent),
		),
		cfg.Property(
			"free",
			"Kernel's measure of free memory.",
			cfg.BytesValue(vmem.Free),
		),
	}, nil
}
