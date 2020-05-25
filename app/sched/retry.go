package sched

import (
	"context"
	"time"

	"github.com/mmcloughlin/goperf/app/db"
	"github.com/mmcloughlin/goperf/app/entity"
)

type retry struct {
	db *db.DB

	maxErrors int
	cooloff   time.Duration
}

// NewRetry builds a scheduler that retries previous error tasks, for which
// there has never been a successful completion. The scheduler will stop
// suggestiing retries after maxErrors total errors for a commit module pair.
// Retries will only be scheduled after the given cooloff period.
func NewRetry(d *db.DB, maxErrors int, cooloff time.Duration) Scheduler {
	return &retry{
		db:        d,
		maxErrors: maxErrors,
		cooloff:   cooloff,
	}
}

func (r *retry) Tasks(ctx context.Context, req *Request) ([]*Task, error) {
	lastAttempt := time.Now().Add(-r.cooloff)
	candidates, err := r.db.ListCommitModuleErrors(ctx, req.Worker, r.maxErrors-1, lastAttempt, req.Num)
	if err != nil {
		return nil, err
	}

	tasks := make([]*Task, len(candidates))
	for i, candidate := range candidates {
		tasks[i] = &Task{
			Priority: r.pri(candidate.NumErrors),
			Spec: entity.TaskSpec{
				CommitSHA:  candidate.CommitSHA,
				Type:       entity.TaskTypeModule,
				TargetUUID: candidate.ModuleUUID,
			},
		}
	}

	return tasks, nil
}

// pri returns the priority for a task with n errors so far.
func (r *retry) pri(n int) float64 {
	// Consider retries high priority at first but decline to 0 when we reach
	// max errors.
	return smoothstep(float64(n),
		1, PriorityHighest,
		float64(r.maxErrors), PriorityMin,
	)
}
