// +build !linux

package platform

import (
	"flag"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/runner"
	"github.com/mmcloughlin/cb/pkg/wrap"
)

type Platform struct {
	base     command.Base
	wrappers []subcommands.Command
}

func New(b command.Base) *Platform {
	return &Platform{base: b}
}

func (p *Platform) Wrappers() []subcommands.Command {
	p.wrappers = []subcommands.Command{
		wrap.NewConfigDefault(p.base),
		wrap.NewPrioritize(p.base),
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
