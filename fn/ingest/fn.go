package ingest

import (
	"context"
	"log"

	"github.com/mmcloughlin/cb/app/gcs"
	"github.com/mmcloughlin/cb/app/results"
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

	// Output.
	for _, r := range rs {
		log.Print(
			r.Benchmark.FullName,
			r.Iterations,
			r.Value,
			r.Benchmark.Unit,
		)
	}

	return nil
}
