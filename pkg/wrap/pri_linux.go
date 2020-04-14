package wrap

import (
	"flag"

	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/pkg/proc"
)

func prioritizeactions(l *zap.Logger) []action {
	return []action{
		&niceaction{},
		&rtaction{log: l},
	}
}

type rtaction struct {
	prio int
	log  *zap.Logger
}

func (a *rtaction) SetFlags(f *flag.FlagSet) {
	f.IntVar(&a.prio, "rt", 0, "if non-zero, attempt to set real-time priority")
}

func (a *rtaction) Apply() error {
	err := proc.SetScheduler(0, proc.SCHED_RR, &proc.SchedParam{Priority: a.prio})
	if err != nil {
		a.log.Error("failed to set realtime priority", zap.Error(err))
	}
	return nil
}
