package ingest

import (
	"context"
	"log"

	"github.com/mmcloughlin/cb/app/gcs"
	"github.com/mmcloughlin/cb/app/results"
	"github.com/mmcloughlin/cb/app/service"
)

// GCSEvent is the payload of a GCS event.
type GCSEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

// Handle GCS event.
func Handle(ctx context.Context, e GCSEvent) error {
	log.Printf("bucket: %v\n", e.Bucket)
	log.Printf("file: %v\n", e.Name)

	// Load results.
	bucket, err := gcs.New(ctx, e.Bucket)
	if err != nil {
		return err
	}

	l, err := results.NewLoader(results.WithFilesystem(bucket))
	if err != nil {
		return err
	}

	rs, err := l.Load(ctx, e.Name)
	if err != nil {
		return err
	}

	// Open database connection.
	d, err := service.DB(ctx)
	if err != nil {
		return err
	}
	defer d.Close()

	// Write to object storage.
	for _, r := range rs {
		if err := d.StoreResult(ctx, r); err != nil {
			return err
		}
		log.Printf("inserted result: %s %v %s", r.Benchmark.FullName, r.Value, r.Benchmark.Unit)
	}

	log.Printf("inserted %d results", len(rs))

	return nil
}
