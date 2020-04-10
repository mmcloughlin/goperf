package db_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcloughlin/cb/app/db/dbtest"
	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/internal/fixture"
)

func TestDBTransitionTaskStatus(t *testing.T) {
	db := dbtest.Open(t)

	// Create task.
	ctx := context.Background()
	task, err := db.CreateTask(ctx, fixture.Worker, fixture.TaskSpec)
	if err != nil {
		t.Fatal(err)
	}

	// Transition to in_progress status.
	err = db.TransitionTaskStatus(ctx, task.UUID, []entity.TaskStatus{entity.TaskStatusCreated}, entity.TaskStatusInProgress)
	if err != nil {
		t.Fatal(err)
	}

	// Fetch the task.
	updated, err := db.FindTaskByUUID(ctx, task.UUID)
	if err != nil {
		t.Fatal(err)
	}

	if updated.Status != entity.TaskStatusInProgress {
		t.Error("uncorrect status")
	}

	if !updated.LastStatusUpdate.After(task.LastStatusUpdate) {
		t.Error("last update timestamp was not changed")
	}
}

func TestDBTransitionTaskStatusNoChange(t *testing.T) {
	db := dbtest.Open(t)

	// Create task.
	ctx := context.Background()
	task, err := db.CreateTask(ctx, fixture.Worker, fixture.TaskSpec)
	if err != nil {
		t.Fatal(err)
	}

	// Transition from completed_success to in_progress. This should fail since it's in created state.
	err = db.TransitionTaskStatus(ctx, task.UUID, []entity.TaskStatus{entity.TaskStatusCompleteSuccess}, entity.TaskStatusInProgress)
	if err == nil {
		t.Fatal("expected error; got nil")
	}

	// Fetch the task. We expect it to be unchanged.
	unchanged, err := db.FindTaskByUUID(ctx, task.UUID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(task, unchanged); diff != "" {
		t.Fatalf("expected task to be unchanged\n%s", diff)
	}
}
