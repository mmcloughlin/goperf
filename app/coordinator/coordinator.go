package coordinator

import (
	"context"
	"sort"
	"time"

	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/sched"
	"github.com/mmcloughlin/cb/internal/errutil"
	"github.com/mmcloughlin/cb/pkg/job"
)

type Coordinator struct {
	db    *db.DB
	sched sched.Scheduler
}

// Jobs requests next jobs for a worker.
func (c *Coordinator) Jobs(ctx context.Context, req *JobsRequest) (*JobsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Determine pending tasks for the worker.
	pending, err := c.db.ListWorkerTasksPending(ctx, req.Worker)
	if err != nil {
		return nil, err
	}

	// Fetch proposed work.
	proposed, err := c.sched.Tasks(ctx, &sched.Request{
		Worker: req.Worker,
		Num:    len(pending) + 1,
	})
	if err != nil {
		return nil, err
	}

	// Select the highest priority task that is not still in a pending state.
	sort.Stable(sort.Reverse(sched.TasksByPriority(proposed)))

	var selected *sched.Task
	for _, task := range proposed {
		if !tasksContainSpec(pending, task.Spec) {
			selected = task
			break
		}
	}

	if selected == nil {
		return NoJobsAvailable(), nil
	}

	// Map to a job definition.
	j, err := c.job(ctx, req, selected.Spec)
	if err != nil {
		return nil, err
	}

	// Create the task and link it with the job.
	t, err := c.db.CreateTask(ctx, req.Worker, selected.Spec)
	if err != nil {
		return nil, err
	}
	j.TaskUUID = t.UUID

	return &JobsResponse{
		Jobs: []*Job{j},
	}, nil
}

// job expands a task specification to a job definition.
func (c *Coordinator) job(ctx context.Context, req *JobsRequest, s entity.TaskSpec) (*Job, error) {
	if !s.Type.IsATaskType() {
		return nil, errutil.AssertionFailure("invalid task type")
	}
	switch s.Type {
	case entity.TaskTypeModule:
		return c.modulejob(ctx, req, s)
	default:
		return nil, errutil.UnhandledCase(s.Type)
	}
}

// modulejob maps a TaskTypeModule task to a job definition.
func (c *Coordinator) modulejob(ctx context.Context, req *JobsRequest, s entity.TaskSpec) (*Job, error) {
	// Lookup the module.
	m, err := c.db.FindModuleByUUID(ctx, s.TargetUUID)
	if err != nil {
		return nil, err
	}

	// TODO(mbm): configurable job defaults
	return &Job{
		CommitSHA: s.CommitSHA,
		Suite: job.Suite{
			Module: job.Module{
				Path:    m.Path,
				Version: m.Version,
			},
			Short:     true,
			BenchTime: time.Second,
		},
	}, nil
}

// tasksContainSpec reports whether any of the tasks have the given spec.
func tasksContainSpec(tasks []*entity.Task, s entity.TaskSpec) bool {
	for _, task := range tasks {
		if task.Spec == s {
			return true
		}
	}
	return false
}
