package cfg

import (
	"github.com/shirou/gopsutil/mem"
)

var Mem Provider = ProviderFunc(memory)

func memory() (Configuration, error) {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	p := NewPrefixed("mem")
	p.Add("total", BytesValue(vmem.Total))
	p.Add("available", BytesValue(vmem.Available))
	p.Add("used", BytesValue(vmem.Used))
	p.Add("free", BytesValue(vmem.Free))
	p.Add("used-percent", PercentageValue(vmem.UsedPercent))

	return p.Configuration()
}
