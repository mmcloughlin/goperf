package coordinator

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/sched"
	"github.com/mmcloughlin/cb/internal/errutil"
	"github.com/mmcloughlin/cb/pkg/cfg"
	"github.com/mmcloughlin/cb/pkg/fs"
	"github.com/mmcloughlin/cb/pkg/job"
)

type Coordinator struct {
	db     *db.DB
	sched  sched.Scheduler
	datafs fs.Writable
	log    *zap.Logger
}

func New(d *db.DB, s sched.Scheduler, w fs.Writable) *Coordinator {
	return &Coordinator{
		db:     d,
		sched:  s,
		datafs: w,
		log:    zap.NewNop(),
	}
}

// SetLogger sets the logger used by the Coordinator.
func (c *Coordinator) SetLogger(l *zap.Logger) {
	c.log = l.Named("coordinator")
}

// Jobs requests next jobs for a worker.
func (c *Coordinator) Jobs(ctx context.Context, req *JobsRequest) (*JobsResponse, error) {
	log := c.log.With(zap.String("worker", req.Worker))
	log.Debug("jobs request")

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Determine pending tasks for the worker.
	pending, err := c.db.ListWorkerTasksPending(ctx, req.Worker)
	if err != nil {
		return nil, err
	}

	log.Debug("found pending tasks", zap.Int("num_pending", len(pending)))

	// Fetch proposed work.
	proposed, err := c.sched.Tasks(ctx, &sched.Request{
		Worker: req.Worker,
		Num:    len(pending) + 1,
	})
	if err != nil {
		return nil, err
	}

	log.Debug("scheduler proposed tasks", zap.Int("num_proposed", len(proposed)))

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
	j, err := c.job(ctx, selected.Spec)
	if err != nil {
		return nil, err
	}

	// Create the task and link it with the job.
	t, err := c.db.CreateTask(ctx, req.Worker, selected.Spec)
	if err != nil {
		return nil, err
	}
	j.UUID = t.UUID

	return &JobsResponse{
		Jobs: []*Job{j},
	}, nil
}

// job expands a task specification to a job definition.
func (c *Coordinator) job(ctx context.Context, s entity.TaskSpec) (*Job, error) {
	if !s.Type.IsATaskType() {
		return nil, errutil.AssertionFailure("invalid task type")
	}
	switch s.Type {
	case entity.TaskTypeModule:
		return c.modulejob(ctx, s)
	default:
		return nil, errutil.UnhandledCase(s.Type)
	}
}

// modulejob maps a TaskTypeModule task to a job definition.
func (c *Coordinator) modulejob(ctx context.Context, s entity.TaskSpec) (*Job, error) {
	// Lookup the module.
	m, err := c.db.FindModuleByUUID(ctx, s.TargetUUID)
	if err != nil {
		return nil, fmt.Errorf("find module: %w", err)
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
			BenchTime: 10 * time.Millisecond,
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

// StatusChange records a job status change.
func (c *Coordinator) StatusChange(ctx context.Context, req *StatusChangeRequest) error {
	log := c.log.With(
		zap.String("worker", req.Worker),
		zap.Stringer("job_uuid", req.UUID),
	)
	log.Debug(fmt.Sprintf("change status to %s", req.To))

	if err := req.Validate(); err != nil {
		return err
	}

	// Find the task.
	task, err := c.findWorkerTask(ctx, req.Worker, req.UUID)
	if err != nil {
		return err
	}

	// Update the state.
	if err := c.db.TransitionTaskStatus(ctx, task.UUID, req.From, req.To); err != nil {
		return err
	}

	return nil
}

// Result processes a datafile upload.
func (c *Coordinator) Result(ctx context.Context, req *ResultRequest) error {
	log := c.log.With(
		zap.String("worker", req.Worker),
		zap.Stringer("job_uuid", req.UUID),
	)
	log.Debug("result upload")

	if err := req.Validate(); err != nil {
		return err
	}

	// Find the task.
	task, err := c.findWorkerTask(ctx, req.Worker, req.UUID)
	if err != nil {
		return fmt.Errorf("find task: %w", err)
	}

	// Record start of upload.
	to := entity.TaskStatusResultUploadStarted
	from := []entity.TaskStatus{
		entity.TaskStatusInProgress,
		to, // allow repeat attempts
	}
	if err := c.db.TransitionTaskStatus(ctx, task.UUID, from, to); err != nil {
		return fmt.Errorf("update task status: %w", err)
	}

	// Write the file.
	log.Debug("writing to filesystem")

	datafile, err := c.write(ctx, req, task)
	if err != nil {
		return fmt.Errorf("results upload: %w", err)
	}

	// Record successful upload.
	log.Debug("record successful upload in database")

	if err := c.db.RecordTaskDataUpload(ctx, task.UUID, datafile); err != nil {
		return err
	}

	return nil
}

// write results file to filesystem.
func (c *Coordinator) write(ctx context.Context, r io.Reader, task *entity.Task) (_ *entity.DataFile, err error) {
	// Create config header.
	config := taskConfig(task)
	hdr := bytes.NewBuffer(nil)
	if err := cfg.Write(hdr, config); err != nil {
		return nil, err
	}

	r = io.MultiReader(hdr, r)

	// Create the file.
	name := task.UUID.String()
	w, err := c.datafs.Create(ctx, name)
	if err != nil {
		return nil, err
	}
	defer errutil.CheckClose(&err, w)

	// Copy to filesystem, hashing as we go.
	h := sha256.New()
	tee := io.TeeReader(r, h)
	if _, err := io.Copy(w, tee); err != nil {
		return nil, err
	}

	// Build datafile object.
	datafile := &entity.DataFile{Name: name}
	h.Sum(datafile.SHA256[:0])

	return datafile, nil
}

// findWorkerTask looks up a task by ID, verifying that it belongs to worker.
func (c *Coordinator) findWorkerTask(ctx context.Context, worker string, id uuid.UUID) (*entity.Task, error) {
	task, err := c.db.FindTaskByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	if task.Worker != worker {
		return nil, errors.New("job does not belong to worker")
	}

	return task, nil
}

// taskConfig generates configuration lines with metadata about the task.
func taskConfig(t *entity.Task) cfg.Configuration {
	return cfg.Configuration{
		cfg.Section(
			"task",
			"task properties",
			cfg.Property("uuid", "task unique identifier", t.UUID),
			cfg.Property("worker", "name of worker that executed the task", cfg.StringValue(t.Worker)),
			cfg.Property("type", "task type", t.Spec.Type),
			cfg.Property("target", "unique identifier of target under test", t.Spec.TargetUUID),
			cfg.Property("commitsha", "commit sha the task was for", cfg.StringValue(t.Spec.CommitSHA)),
		),
	}
}
