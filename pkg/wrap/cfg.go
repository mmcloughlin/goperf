package wrap

import (
	"flag"
	"os"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/goperf/internal/flags"
	"github.com/mmcloughlin/goperf/pkg/cfg"
	"github.com/mmcloughlin/goperf/pkg/command"
)

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
