package wrap

import (
	"flag"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/proc"
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

type cpusetaction struct {
	name string
}

func (a *cpusetaction) SetFlags(f *flag.FlagSet) {
	f.StringVar(&a.name, "name", "", "cpuset to run in")
}

func (a *cpusetaction) Apply() error {
	return proc.SetCPUSetSelf(a.name)
}
