package sys

import (
	cpuutil "github.com/shirou/gopsutil/cpu"

	"github.com/mmcloughlin/cb/internal/errutil"
	"github.com/mmcloughlin/cb/pkg/cfg"
	"github.com/mmcloughlin/cb/pkg/proc"
)

var AffineCPU = cfg.NewProviderFunc("affinecpu", "CPU information for processors assigned to this process", affinecpu)

func affinecpu() (cfg.Configuration, error) {
	procs, err := cpuutil.Info()
	if err != nil {
		return nil, err
	}

	affinty, err := proc.AffinitySelf()
	if err != nil {
		return nil, err
	}

	c := cfg.Configuration{}
	for idx, cpu := range affinty {
		proc := procs[cpu]
		if int(proc.CPU) != cpu {
			return nil, errutil.AssertionFailure("unexpected cpu index")
		}
		c = append(c, processor(proc, idx))
	}
	return c, nil
}
