package sched

import (
	"context"
	"time"

	"github.com/mmcloughlin/cb/app/db"
)

type recent struct {
	db     *db.DB
	window time.Duration
}

func NewRecentCommits(d *db.DB, window time.Duration) Scheduler {
	return &recent{
		db:     d,
		window: window,
	}
}

func (r *recent) Tasks(ctx context.Context, req *Request) ([]*Task, error) {
	since := time.Now().Add(-r.window)
	specs, err := r.db.ListTaskSpecsRecentCommitsWithoutWorkerResults(ctx, req.Worker, since, req.Num)
	if err != nil {
		return nil, err
	}

	tasks := make([]*Task, len(specs))
	for i, spec := range specs {
		tasks[i] = &Task{
			Priority: 1.0,
			Spec:     spec,
		}
	}

	return tasks, nil
}
