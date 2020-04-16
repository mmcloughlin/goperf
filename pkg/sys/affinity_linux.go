package sys

import (
	"fmt"
	"strconv"

	"github.com/mmcloughlin/cb/pkg/cfg"
	"github.com/mmcloughlin/cb/pkg/proc"
)

var AffineCPU = cfg.NewProviderFunc("affinecpu", "CPU information for processors assigned to this process", affinecpu)

func affinecpu() (cfg.Configuration, error) {
	procs, err := cpuinfo("/proc/cpuinfo", cfg.TagPerfCritical)
	if err != nil {
		return nil, err
	}

	affinty, err := proc.AffinitySelf()
	if err != nil {
		return nil, err
	}

	c := cfg.Configuration{}
	for idx, cpu := range affinty {
		section := cfg.Section(
			cfg.Key("cpu"+strconv.Itoa(idx)),
			fmt.Sprintf("information about processor %d assigned to this process", idx),
			procs[cpu]...,
		)
		c = append(c, section)
	}
	return c, nil
}
