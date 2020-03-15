package wrap

import (
	"github.com/mmcloughlin/cb/pkg/cfg"
	"github.com/mmcloughlin/cb/pkg/proc"
	"github.com/mmcloughlin/cb/pkg/sys"
)

// DefaultProviders is the default list of config providers on Linux.
var DefaultProviders = cfg.Providers{
	sys.Host,
	sys.LoadAverage,
	sys.VirtualMemory,
	sys.CPU,
	sys.AffineCPU,
	sys.Caches{},
	sys.CPUFreq{},
	sys.IntelPState{},
	proc.Stat{},
}
