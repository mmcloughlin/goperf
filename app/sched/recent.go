package sched

import (
	"context"
	"time"

	"github.com/mmcloughlin/goperf/app/db"
	"github.com/mmcloughlin/goperf/app/entity"
)

type recent struct {
	db  *db.DB
	pri TimePriority
}

// NewRecentCommits builds a scheduler that proposes tasks for recent comments
// and modules with no completed tasks in the database. Priority is computed
// with the commit time and supplied priority function.
func NewRecentCommits(d *db.DB, pri TimePriority) Scheduler {
	return &recent{
		db:  d,
		pri: pri,
	}
}

func (r *recent) Tasks(ctx context.Context, req *Request) ([]*Task, error) {
	cms, err := r.db.ListCommitModulesWithoutCompleteTasks(ctx, req.Worker, req.Num)
	if err != nil {
		return nil, err
	}

	tasks := make([]*Task, len(cms))
	for i, cm := range cms {
		tasks[i] = &Task{
			Priority: r.pri(cm.CommitTime),
			Spec: entity.TaskSpec{
				CommitSHA:  cm.CommitSHA,
				Type:       entity.TaskTypeModule,
				TargetUUID: cm.ModuleUUID,
			},
		}
	}

	return tasks, nil
}

// TimePriority is a method of determining priority based on a time.
type TimePriority func(time.Time) float64

// ConstantTimePriority returns a function that returns p for all times.
func ConstantTimePriority(p float64) TimePriority {
	return func(t time.Time) float64 { return p }
}

// TimeSinceSmoothStep has priority p0 for times up to d0 from now, priority p1
// for times over d1 from now and smoothly transitions between the two.
func TimeSinceSmoothStep(d0 time.Duration, p0 float64, d1 time.Duration, p1 float64) TimePriority {
	x0 := float64(d0)
	x1 := float64(d1)
	return func(t time.Time) float64 {
		x := float64(time.Since(t))
		return smoothstep(x, x0, p0, x1, p1)
	}
}
