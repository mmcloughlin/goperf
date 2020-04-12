package sched

import (
	"context"

	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/app/entity"
)

type recent struct {
	db *db.DB
}

func NewRecentCommits(d *db.DB) Scheduler {
	return &recent{db: d}
}

func (r *recent) Tasks(ctx context.Context, req *Request) ([]*Task, error) {
	cms, err := r.db.ListCommitModulesWithoutCompleteTasks(ctx, req.Worker, req.Num)
	if err != nil {
		return nil, err
	}

	tasks := make([]*Task, len(cms))
	for i, cm := range cms {
		tasks[i] = &Task{
			Spec: entity.TaskSpec{
				CommitSHA:  cm.CommitSHA,
				Type:       entity.TaskTypeModule,
				TargetUUID: cm.ModuleUUID,
			},
		}
	}

	return tasks, nil
}
