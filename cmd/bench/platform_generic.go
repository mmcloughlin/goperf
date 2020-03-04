// +build !linux

package main

import (
	"flag"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/runner"
	"github.com/mmcloughlin/cb/pkg/wrap"
)

type Platform struct {
	wrappers []subcommands.Command
}

func (p *Platform) Wrappers(b command.Base) []subcommands.Command {
	p.wrappers = []subcommands.Command{
		wrap.NewConfigDefault(b),
		wrap.NewPrioritize(b),
	}
	return p.wrappers
}

func (Platform) SetFlags(f *flag.FlagSet) {}

// ConfigureRunner sets benchmark runner options.
func (p *Platform) ConfigureRunner(r *runner.Runner) error {
	for _, wrapper := range p.wrappers {
		w, err := wrap.RunUnder(wrapper)
		if err != nil {
			return err
		}
		r.Wrap(w)
	}
	return nil
}
