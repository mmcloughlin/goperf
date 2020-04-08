// Package sched provides methods of scheduling tasks.
package sched

import (
	"context"

	"github.com/mmcloughlin/cb/app/entity"
)

// Request for work.
type Request struct {
	Worker string // worker to request for
	Num    int    // request this many proposed tasks
}

// Task is prioritized work proposed by a scheduler.
type Task struct {
	Priority float64
	Spec     entity.TaskSpec
}

// NewTask builds a task with the supplied priority and specifiction.
func NewTask(pri float64, s entity.TaskSpec) *Task {
	return &Task{
		Priority: pri,
		Spec:     s,
	}
}

// TasksByPriority provides a sort.Interface for sorting tasks in increasing priority.
type TasksByPriority []*Task

func (t TasksByPriority) Len() int           { return len(t) }
func (t TasksByPriority) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t TasksByPriority) Less(i, j int) bool { return t[i].Priority < t[j].Priority }

// Scheduler proposes work.
type Scheduler interface {
	Tasks(ctx context.Context, req *Request) ([]*Task, error)
}

// SchedulerFunc adapts a function to the Scheduler interface.
type SchedulerFunc func(ctx context.Context, req *Request) ([]*Task, error)

// Tasks calls f.
func (f SchedulerFunc) Tasks(ctx context.Context, req *Request) ([]*Task, error) {
	return f(ctx, req)
}

// StaticScheduler always returns the same tasks.
func StaticScheduler(tasks []*Task) Scheduler {
	return SchedulerFunc(func(ctx context.Context, req *Request) ([]*Task, error) {
		return tasks, nil
	})
}

// SingleTaskScheduler always returns the given task.
func SingleTaskScheduler(task *Task) Scheduler {
	return StaticScheduler([]*Task{task})
}
