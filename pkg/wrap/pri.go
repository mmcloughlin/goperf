package wrap

import (
	"flag"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/proc"
)

func NewPrioritize(b command.Base) subcommands.Command {
	return &wrapper{
		Base:     b,
		name:     "pri",
		synopsis: "prioritize a process",
		actions:  prioritizeactions(b.Log),
	}
}

type niceaction struct {
	prio int
}

func (a *niceaction) SetFlags(f *flag.FlagSet) {
	f.IntVar(&a.prio, "nice", -20, "nice value")
}

func (a *niceaction) Apply() error {
	return proc.SetPriority(0, a.prio)
}
