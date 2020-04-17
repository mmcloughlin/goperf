package wrap

import (
	"github.com/mmcloughlin/cb/meta"
	"github.com/mmcloughlin/cb/pkg/cfg"
	"github.com/mmcloughlin/cb/pkg/gocfg"
	"github.com/mmcloughlin/cb/pkg/proc"
	"github.com/mmcloughlin/cb/pkg/sys"
)

// DefaultProviders is the default list of config providers on Linux.
var DefaultProviders = cfg.Providers{
	meta.Provider{},
	gocfg.Provider,
	sys.Host,
	sys.LoadAverage,
	sys.VirtualMemory,
	sys.Thermal{},
	sys.CPU,
	sys.AffineCPU,
	sys.Caches(),
	sys.AffineCaches(),
	sys.CPUFreq(),
	sys.AffineCPUFreq(),
	sys.IntelPState{},
	sys.SMT{},
	proc.Stat{},
}
