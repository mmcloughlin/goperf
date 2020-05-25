package wrap

import (
	"github.com/mmcloughlin/goperf/meta"
	"github.com/mmcloughlin/goperf/pkg/cfg"
	"github.com/mmcloughlin/goperf/pkg/gocfg"
	"github.com/mmcloughlin/goperf/pkg/proc"
	"github.com/mmcloughlin/goperf/pkg/sys"
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
