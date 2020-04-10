package coordinator_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/coordinator"
	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/app/db/dbtest"
	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/internal/fixture"
	"github.com/mmcloughlin/cb/app/sched"
	"github.com/mmcloughlin/cb/pkg/job"
	"github.com/mmcloughlin/cb/pkg/lg"
)

type Integration struct {
	DB *db.DB

	ctx    context.Context
	server *httptest.Server
}

func NewIntegration(t *testing.T) *Integration {
	// Run in parallel to confirm multiple workers act on the database
	// independently.
	t.Parallel()

	ctx := context.Background()
	db := dbtest.Open(t)
	l := lg.Test(t)

	// Ensure the module is in the database.
	if err := db.StoreModule(ctx, fixture.Module); err != nil {
		t.Fatal(err)
	}

	// Create coordinator server.
	scheduler := sched.SingleTaskScheduler(sched.NewTask(0, fixture.TaskSpec))
	c := coordinator.New(db, scheduler)
	c.SetLogger(l)
	h := coordinator.NewHandlers(c, l)

	s := httptest.NewServer(h)

	return &Integration{
		DB:     db,
		ctx:    ctx,
		server: s,
	}
}

func (i *Integration) Context() context.Context { return i.ctx }

func (i *Integration) Close() {
	i.server.Close()
}

func (i *Integration) NewClient(worker string) *coordinator.Client {
	return coordinator.NewClient(http.DefaultClient, i.server.URL, worker)
}

func TestIntegrationJobCreation(t *testing.T) {
	i := NewIntegration(t)
	ctx := i.Context()
	worker := "test_job_creation"
	client := i.NewClient(worker)

	// Request work.
	jobs := []*coordinator.Job{}

	for i := 0; i < 3; i++ {
		res, err := client.Jobs(ctx)
		if err != nil {
			t.Fatal(err)
		}
		jobs = append(jobs, res.Jobs...)
	}

	// Only the first one should have returned anything, since the worker didn't
	// return a result.
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job; got %d", len(jobs))
	}

	j := jobs[0]
	expect := &coordinator.Job{
		UUID:      j.UUID, // can't predict this
		CommitSHA: fixture.TaskSpec.CommitSHA,
		Suite: job.Suite{
			Module: job.Module{
				Path:    fixture.Module.Path,
				Version: fixture.Module.Version,
			},
			Short:     true,
			BenchTime: time.Second,
		},
	}

	if diff := cmp.Diff(expect, j); diff != "" {
		t.Fatalf("job mismatch\n%s", diff)
	}

	// Verify the task was added to the database.
	got, err := i.DB.FindTaskByUUID(ctx, j.UUID)
	if err != nil {
		t.Fatalf("could not find task in the database: %v", err)
	}

	expecttask := &entity.Task{
		UUID:             j.UUID,
		Worker:           worker,
		Spec:             fixture.TaskSpec,
		Status:           entity.TaskStatusCreated,
		LastStatusUpdate: got.LastStatusUpdate,
		DatafileUUID:     uuid.Nil,
	}

	if diff := cmp.Diff(expecttask, got); diff != "" {
		t.Fatalf("task mismatch\n%s", diff)
	}
}

func TestIntegrationJobStart(t *testing.T) {
	i := NewIntegration(t)
	ctx := i.Context()
	worker := "test_job_start"
	client := i.NewClient(worker)

	// Request work.
	res, err := client.Jobs(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Jobs) != 1 {
		t.Fatalf("expected 1 job; got %d", len(res.Jobs))
	}
	j := res.Jobs[0]

	// Start it.
	if err := client.Start(ctx, j.UUID); err != nil {
		t.Fatal(err)
	}

	// Verify the task has the right status.
	task, err := i.DB.FindTaskByUUID(ctx, j.UUID)
	if err != nil {
		t.Fatalf("could not find task in the database: %v", err)
	}

	if task.Status != entity.TaskStatusInProgress {
		t.Fatalf("expected task to be in progress; got %s", task.Status)
	}
}
