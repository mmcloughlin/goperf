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

	rt   int
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
	f.IntVar(&cmd.rt, "rt", 99, "if non-zero, attempt to set realtime priority")
	f.IntVar(&cmd.nice, "nice", -20, "nice value, used as a fallback if realtime fails")
}

func (cmd *Prioritize) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// Sub-process arguments.
	args := f.Args()

	// Apply process priority.
	runtime.LockOSThread()

	nice := true
	if cmd.rt != 0 {
		err := proc.SetScheduler(0, proc.SCHED_RR, &proc.SchedParam{Priority: cmd.rt})
		if err != nil {
			cmd.Log.Printf("failed to set realtime priority")
		} else {
			nice = false
		}
	}

	if nice {
		if err := proc.SetPriority(0, cmd.nice); err != nil {
			return cmd.Error(err)
		}
	}

	// Execute the sub-process.
	return cmd.Status(proc.Exec(args))
}
