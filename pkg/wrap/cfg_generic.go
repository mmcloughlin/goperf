// +build !linux

package wrap

import (
	"github.com/mmcloughlin/goperf/meta"
	"github.com/mmcloughlin/goperf/pkg/cfg"
	"github.com/mmcloughlin/goperf/pkg/gocfg"
	"github.com/mmcloughlin/goperf/pkg/sys"
)

// DefaultProviders is the default list of config providers for generic platforms.
var DefaultProviders = cfg.Providers{
	meta.Provider{},
	gocfg.Provider,
	sys.Host,
	sys.LoadAverage,
	sys.VirtualMemory,
	sys.CPU,
}
