package wrap

import (
	"flag"

	"github.com/mmcloughlin/cb/pkg/lg"
	"github.com/mmcloughlin/cb/pkg/proc"
)

func prioritizeactions(l lg.Logger) []action {
	return []action{
		&niceaction{},
		&rtaction{l: l},
	}
}

type rtaction struct {
	prio int
	l    lg.Logger
}

func (a *rtaction) SetFlags(f *flag.FlagSet) {
	f.IntVar(&a.prio, "rt", 99, "if non-zero, attempt to set real-time priority")
}

func (a *rtaction) Apply() error {
	err := proc.SetScheduler(0, proc.SCHED_RR, &proc.SchedParam{Priority: a.prio})
	if err != nil {
		a.l.Printf("failed to set realtime priority")
	}
	return nil
}
