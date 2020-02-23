package wrap

import (
	"context"
	"flag"
	"os"
	"runtime"

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

type Config struct {
	command.Base

	providers cfg.Providers

	keys flags.Strings
}

func NewConfig(b command.Base, providers cfg.Providers) *Config {
	return &Config{
		Base:      b,
		providers: providers,
		keys:      providers.Keys(),
	}
}

func NewConfigDefault(b command.Base) *Config {
	return NewConfig(b, DefaultProviders)
}

func (*Config) Name() string { return "cfg" }

func (*Config) Synopsis() string {
	return "execute a process with additional environmental configuration output"
}

func (*Config) Usage() string {
	return ``
}

func (cmd *Config) SetFlags(f *flag.FlagSet) {
	f.Var(&cmd.keys, "cfg", "config types to include")
}

func (cmd *Config) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// Determine the config providers.
	ps, err := cmd.providers.Select(cmd.keys...)
	if err != nil {
		return cmd.Error(err)
	}

	args := f.Args()

	// Read configuration.
	runtime.LockOSThread()

	c, err := ps.Configuration()
	if err != nil {
		return cmd.Error(err)
	}

	// Output.
	if err := cfg.Write(os.Stdout, c); err != nil {
		return cmd.Error(err)
	}

	// Execute the sub-process.
	return cmd.Status(proc.Exec(args))
}
