package db

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/db/internal/db"
	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/internal/errutil"
)

// CreateTask creates a new task.
func (d *DB) CreateTask(ctx context.Context, worker string, s entity.TaskSpec) (*entity.Task, error) {
	var t *entity.Task
	err := d.tx(ctx, func(q *db.Queries) error {
		var err error
		t, err = createTask(ctx, q, worker, s)
		return err
	})
	return t, err
}

func createTask(ctx context.Context, q *db.Queries, worker string, s entity.TaskSpec) (*entity.Task, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	sha, err := hex.DecodeString(s.CommitSHA)
	if err != nil {
		return nil, fmt.Errorf("invalid sha: %w", err)
	}

	typ, err := toTaskType(s.Type)
	if err != nil {
		return nil, err
	}

	t, err := q.CreateTask(ctx, db.CreateTaskParams{
		UUID:       id,
		Worker:     worker,
		CommitSHA:  sha,
		Type:       typ,
		TargetUUID: s.TargetUUID,
	})
	if err != nil {
		return nil, err
	}

	return mapTask(t)
}

// ListWorkerTasksPending returns tasks assigned to a worker in a pending state.
func (d *DB) ListWorkerTasksPending(ctx context.Context, worker string) ([]*entity.Task, error) {
	return d.ListWorkerTasksWithStatus(ctx, worker, entity.TaskStatusPendingValues())
}

// ListWorkerTasksWithStatus returns tasks assigned to a worker in the given states.
func (d *DB) ListWorkerTasksWithStatus(ctx context.Context, worker string, statuses []entity.TaskStatus) ([]*entity.Task, error) {
	var ts []*entity.Task
	err := d.tx(ctx, func(q *db.Queries) error {
		var err error
		ts, err = listWorkerTasksWithSpecAndStatus(ctx, q, worker, statuses)
		return err
	})
	return ts, err
}

func listWorkerTasksWithSpecAndStatus(ctx context.Context, q *db.Queries, worker string, statuses []entity.TaskStatus) ([]*entity.Task, error) {
	taskStatuses, err := toTaskStatuses(statuses)
	if err != nil {
		return nil, err
	}

	ts, err := q.WorkerTasksWithStatus(ctx, db.WorkerTasksWithStatusParams{
		Worker:   worker,
		Statuses: taskStatuses,
	})
	if err != nil {
		return nil, err
	}

	output := make([]*entity.Task, len(ts))
	for i, t := range ts {
		output[i], err = mapTask(t)
		if err != nil {
			return nil, err
		}
	}

	return output, nil
}

// toTaskType maps a task type to the corresponding database enum value.
func toTaskType(t entity.TaskType) (db.TaskType, error) {
	if !t.IsATaskType() {
		return "", errutil.AssertionFailure("invalid task type")
	}
	switch t {
	case entity.TaskTypeModule:
		return db.TaskTypeModule, nil
	default:
		return "", errutil.UnhandledCase(t)
	}
}

func toTaskStatuses(statuses []entity.TaskStatus) ([]db.TaskStatus, error) {
	ss := make([]db.TaskStatus, 0, len(statuses))
	for _, status := range statuses {
		s, err := toTaskStatus(status)
		if err != nil {
			return nil, err
		}
		ss = append(ss, s)
	}
	return ss, nil
}

func toTaskStatus(status entity.TaskStatus) (db.TaskStatus, error) {
	if !status.IsATaskStatus() {
		return "", errutil.AssertionFailure("invalid task status")
	}
	switch status {
	case entity.TaskStatusCreated:
		return db.TaskStatusCreated, nil
	case entity.TaskStatusInProgress:
		return db.TaskStatusInProgress, nil
	case entity.TaskStatusCompleteSuccess:
		return db.TaskStatusCompleteSuccess, nil
	case entity.TaskStatusCompleteError:
		return db.TaskStatusCompleteError, nil
	default:
		return "", errutil.UnhandledCase(status)
	}
}

func mapTask(t db.Task) (*entity.Task, error) {
	typ, err := mapTaskType(t.Type)
	if err != nil {
		return nil, err
	}
	status, err := mapTaskStatus(t.Status)
	if err != nil {
		return nil, err
	}
	return &entity.Task{
		UUID:   t.UUID,
		Worker: t.Worker,
		Spec: entity.TaskSpec{
			Type:       typ,
			TargetUUID: t.TargetUUID,
			CommitSHA:  hex.EncodeToString(t.CommitSHA),
		},
		Status:           status,
		LastStatusUpdate: t.LastStatusUpdate,
		DatafileUUID:     t.DatafileUUID,
	}, nil
}

func mapTaskType(t db.TaskType) (entity.TaskType, error) {
	switch t {
	case db.TaskTypeModule:
		return entity.TaskTypeModule, nil
	default:
		return 0, errutil.UnhandledCase(t)
	}
}

func mapTaskStatus(status db.TaskStatus) (entity.TaskStatus, error) {
	switch status {
	case db.TaskStatusCreated:
		return entity.TaskStatusCreated, nil
	case db.TaskStatusInProgress:
		return entity.TaskStatusInProgress, nil
	case db.TaskStatusCompleteSuccess:
		return entity.TaskStatusCompleteSuccess, nil
	case db.TaskStatusCompleteError:
		return entity.TaskStatusCompleteError, nil
	default:
		return 0, errutil.UnhandledCase(status)
	}
}
