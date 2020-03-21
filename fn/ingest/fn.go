package ingest

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"

	"github.com/mmcloughlin/cb/app/gcs"
	"github.com/mmcloughlin/cb/app/mapper"
	"github.com/mmcloughlin/cb/app/obj"
	"github.com/mmcloughlin/cb/app/results"
)

// Parameters.
var project = os.Getenv("CB_PROJECT_ID")

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

	// Create Firestore object store.
	fsc, err := firestore.NewClient(ctx, project)
	if err != nil {
		return err
	}
	defer fsc.Close()

	s := obj.OnceSetter(obj.NewFirestore(fsc))

	// Write to object storage.
	for _, r := range rs {
		for _, m := range mapper.ResultsModels(r) {
			if err := s.Set(ctx, m); err != nil {
				return err
			}
		}

		log.Print(r.Benchmark.FullName)
	}

	return nil
}
