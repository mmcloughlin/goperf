// +build !linux

package wrap

import (
	"github.com/mmcloughlin/cb/pkg/cfg"
	"github.com/mmcloughlin/cb/pkg/sys"
)

// DefaultProviders is the default list of config providers for generic platforms.
var DefaultProviders = cfg.Providers{
	sys.Host,
	sys.LoadAverage,
	sys.VirtualMemory,
	sys.CPU,
}
