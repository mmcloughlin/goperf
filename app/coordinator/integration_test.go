package coordinator_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
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
	"github.com/mmcloughlin/cb/internal/test"
	"github.com/mmcloughlin/cb/pkg/fs"
	"github.com/mmcloughlin/cb/pkg/job"
	"github.com/mmcloughlin/cb/pkg/lg"
	"github.com/mmcloughlin/cb/pkg/parse"
)

type Integration struct {
	DB      *db.DB
	DataDir string

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
	dir := test.TempDir(t)
	datafs := fs.NewLocal(dir)
	c := coordinator.New(db, scheduler, datafs)
	c.SetLogger(l)
	h := coordinator.NewHandlers(c, l)
	s := httptest.NewServer(h)
	t.Cleanup(s.Close)

	return &Integration{
		DB:      db,
		DataDir: dir,
		ctx:     ctx,
		server:  s,
	}
}

func (i *Integration) Context() context.Context { return i.ctx }

func (i *Integration) NewClient(worker string) *coordinator.Client {
	return coordinator.NewClient(http.DefaultClient, i.server.URL, worker)
}

func TestIntegrationJobCreation(t *testing.T) {
	i := NewIntegration(t)
	ctx := i.Context()
	worker := "test-job-creation"
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
	worker := "test-job-start"
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

func TestIntegrationJobResultUpload(t *testing.T) {
	i := NewIntegration(t)
	ctx := i.Context()
	worker := "test-result-upload"
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

	// Upload result.
	expect, err := ioutil.ReadFile("testdata/result.txt")
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(expect)
	if err := client.UploadResult(ctx, j.UUID, buf); err != nil {
		t.Fatal(err)
	}

	// Check it was written to the filesystem.
	got, err := ioutil.ReadFile(filepath.Join(i.DataDir, j.UUID.String()))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.HasSuffix(got, expect) {
		t.Fatal("upload mismatch")
	}

	if _, err := parse.Bytes(got); err != nil {
		t.Fatalf("parse upload: %v", err)
	}

	// Confirm database bookeeping.
	task, err := i.DB.FindTaskByUUID(ctx, j.UUID)
	if err != nil {
		t.Fatalf("could not find task in the database: %v", err)
	}

	if task.Status != entity.TaskStatusResultUploaded {
		t.Fatalf("expected task to have result uploaded status; got %s", task.Status)
	}

	if task.DatafileUUID == uuid.Nil {
		t.Fatal("nil data file id")
	}

	f, err := i.DB.FindDataFileByUUID(ctx, task.DatafileUUID)
	if err != nil {
		t.Fatal("could not find corresponding datafile")
	}

	expecthash := sha256.Sum256(got)
	if f.SHA256 != expecthash {
		t.Fatal("sha256 mismatch")
	}
}
