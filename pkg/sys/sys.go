package sys

import (
	"github.com/shirou/gopsutil/mem"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

var VirtualMemory cfg.Provider = cfg.ProviderFunc(virtualmemory)

func virtualmemory() (cfg.Configuration, error) {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	p := cfg.NewPrefixed("virtual-mem")
	p.Add("total", cfg.BytesValue(vmem.Total))
	p.Add("available", cfg.BytesValue(vmem.Available))
	p.Add("used", cfg.BytesValue(vmem.Used))
	p.Add("free", cfg.BytesValue(vmem.Free))
	p.Add("used-percent", cfg.PercentageValue(vmem.UsedPercent))

	return p.Configuration()
}
