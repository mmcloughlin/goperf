package wrap

import (
	"flag"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/proc"
	"github.com/mmcloughlin/cb/pkg/runner"
)

func NewCPUSet(b command.Base) subcommands.Command {
	return &wrapper{
		Base:     b,
		name:     "cpuset",
		synopsis: "execute a process in a given cpuset",
		actions: []action{
			&cpusetaction{},
		},
	}
}

func RunUnderCPUSet(cmd subcommands.Command, name string) (runner.Wrapper, error) {
	return RunUnder(cmd, "-name", name)
}

type cpusetaction struct {
	name string
}

func (a *cpusetaction) SetFlags(f *flag.FlagSet) {
	f.StringVar(&a.name, "name", "", "cpuset name")
}

func (a *cpusetaction) Apply() error {
	return proc.SetCPUSetSelf(a.name)
}
