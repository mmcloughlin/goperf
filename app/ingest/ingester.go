// Package ingest implements ingestion of benchmark results.
package ingest

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/app/db"
	"github.com/mmcloughlin/goperf/app/entity"
	"github.com/mmcloughlin/goperf/app/results"
	"github.com/mmcloughlin/goperf/internal/errutil"
)

// Ingester loads task results files and inserts them in to the database.
type Ingester struct {
	db     *db.DB
	loader *results.Loader
	log    *zap.Logger
}

// New builds an ingester.
func New(d *db.DB, l *results.Loader) *Ingester {
	return &Ingester{
		db:     d,
		loader: l,
		log:    zap.NewNop(),
	}
}

// SetLogger configures Ingester logging.
func (i *Ingester) SetLogger(l *zap.Logger) { i.log = l.Named("ingester") }

// Task ingests results for the given task.
func (i *Ingester) Task(ctx context.Context, id uuid.UUID) error {
	// Find the task.
	task, err := i.db.FindTaskByUUID(ctx, id)
	if err != nil {
		return fmt.Errorf("find task: %w", err)
	}

	if task.Status != entity.TaskStatusResultUploaded {
		return fmt.Errorf("task has status %s", task.Status)
	}

	// Lookup the corresponding datafile.
	f, err := i.db.FindDataFileByUUID(ctx, task.DatafileUUID)
	if err != nil {
		return fmt.Errorf("find data file: %w", err)
	}

	i.log.Info("fetched datafile",
		zap.String("name", f.Name),
		zap.String("sha256", hex.EncodeToString(f.SHA256[:])),
	)

	// Load results.
	rs, err := i.loader.Load(ctx, f.Name)
	if err != nil {
		return fmt.Errorf("load results: %w", err)
	}

	// Sanity check.
	for _, r := range rs {
		if r.File.SHA256 != f.SHA256 {
			return errutil.AssertionFailure("data file hash mismatch")
		}
	}

	// Write to storage.
	if err := i.db.StoreResults(ctx, rs); err != nil {
		return err
	}

	i.log.Debug("inserted results", zap.Int("num_results", len(rs)))

	// Record successful ingestion.
	from := []entity.TaskStatus{entity.TaskStatusResultUploaded}
	to := entity.TaskStatusCompleteSuccess
	if err := i.db.TransitionTaskStatus(ctx, task.UUID, from, to); err != nil {
		return err
	}

	return nil
}
