package ingest

import (
	"context"
	"fmt"
	"log"
	"path"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/app/gcs"
	"github.com/mmcloughlin/cb/app/ingest"
	"github.com/mmcloughlin/cb/app/results"
	"github.com/mmcloughlin/cb/app/service"
)

// Initialization.
var (
	logger   *zap.Logger
	database *db.DB
)

func init() {
	var err error
	ctx := context.Background()

	logger, err = service.Logger()
	if err != nil {
		log.Fatal(err)
	}

	database, err = service.DB(ctx, logger)
}

// GCSEvent is the payload of a GCS event.
type GCSEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

// Handle GCS event.
func Handle(ctx context.Context, e GCSEvent) error {
	logger.Info("received cloud storage trigger",
		zap.String("bucket", e.Bucket),
		zap.String("name", e.Name),
	)

	// Extract task ID from the object name.
	id, err := uuid.Parse(path.Base(e.Name))
	if err != nil {
		return fmt.Errorf("parse task uuid from name: %w", err)
	}

	// Construct Ingester.
	bucket, err := gcs.New(ctx, e.Bucket)
	if err != nil {
		return err
	}

	loader, err := results.NewLoader(results.WithFilesystem(bucket))
	if err != nil {
		return err
	}

	i := ingest.New(database, loader)
	i.SetLogger(logger)

	// Ingest task.
	if err := i.Task(ctx, id); err != nil {
		return fmt.Errorf("task ingest: %w", err)
	}

	return nil
}
