package wrap

import (
	"context"
	"flag"
	"runtime"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/proc"
)

type Prioritize struct {
	command.Base

	nice int
}

func NewPrioritize(b command.Base) *Prioritize {
	return &Prioritize{
		Base: b,
	}
}

func (*Prioritize) Name() string { return "pri" }

func (*Prioritize) Synopsis() string {
	return "prioritize a process"
}

func (*Prioritize) Usage() string {
	return ``
}

func (cmd *Prioritize) SetFlags(f *flag.FlagSet) {
	f.IntVar(&cmd.nice, "nice", -20, "nice value")
}

func (cmd *Prioritize) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// Sub-process arguments.
	args := f.Args()

	// Apply process priority.
	runtime.LockOSThread()

	if err := proc.SetPriority(0, cmd.nice); err != nil {
		return cmd.Error(err)
	}

	// Execute the sub-process.
	return cmd.Status(proc.Exec(args))
}
