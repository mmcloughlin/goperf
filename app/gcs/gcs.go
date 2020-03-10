// Package gcs is a filesystem implementation backed by Google Cloud Storage.
package gcs

import (
	"context"
	"io"

	"cloud.google.com/go/storage"

	"github.com/mmcloughlin/cb/pkg/fs"
)

// gcs is a filesystem implementation backed by a Google Cloud Storage bucket.
type gcs struct {
	bucket *storage.BucketHandle
}

// New builds a filesystem backed by the given Google Cloud Storage bucket.
func New(ctx context.Context, bucket string) (fs.Interface, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &gcs{
		bucket: client.Bucket(bucket),
	}, nil
}

// Create named object for writing.
func (g *gcs) Create(ctx context.Context, name string) (io.WriteCloser, error) {
	return g.bucket.Object(name).NewWriter(ctx), nil
}
