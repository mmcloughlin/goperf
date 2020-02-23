package wrap

import (
	"flag"
	"os"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/internal/flags"
	"github.com/mmcloughlin/cb/pkg/cfg"
	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/proc"
	"github.com/mmcloughlin/cb/pkg/sys"
)

// DefaultProviders is the default list of config providers.
var DefaultProviders = cfg.Providers{
	sys.Host,
	sys.LoadAverage,
	sys.VirtualMemory,
	sys.CPU,
	sys.Caches{},
	sys.CPUFreq{},
	sys.IntelPState{},
	proc.Stat{},
}

func NewConfig(b command.Base, providers cfg.Providers) subcommands.Command {
	a := &configaction{
		providers: providers,
		keys:      providers.Keys(),
	}
	return &wrapper{
		Base:     b,
		name:     "cfg",
		synopsis: "execute a process with additional environmental configuration output",
		actions:  []action{a},
	}
}

func NewConfigDefault(b command.Base) subcommands.Command {
	return NewConfig(b, DefaultProviders)
}

type configaction struct {
	providers cfg.Providers

	keys flags.Strings
}

func (a *configaction) SetFlags(f *flag.FlagSet) {
	f.Var(&a.keys, "cfg", "config types to include")
}

func (a *configaction) Apply() error {
	// Determine the config providers.
	ps, err := a.providers.Select(a.keys...)
	if err != nil {
		return err
	}

	c, err := ps.Configuration()
	if err != nil {
		return err
	}

	// Output.
	if err := cfg.Write(os.Stdout, c); err != nil {
		return err
	}

	return nil
}
