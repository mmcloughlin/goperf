// Package gcs is a filesystem implementation backed by Google Cloud Storage.
package gcs

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"

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

// Remove named object.
func (g *gcs) Remove(ctx context.Context, name string) error {
	return g.bucket.Object(name).Delete(ctx)
}

// Open named object for reading.
func (g *gcs) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	return g.bucket.Object(name).NewReader(ctx)
}

// List objects in bucket.
func (g *gcs) List(ctx context.Context) ([]*fs.FileInfo, error) {
	var files []*fs.FileInfo
	it := g.bucket.Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		files = append(files, &fs.FileInfo{
			Path:    attrs.Name,
			Size:    attrs.Size,
			ModTime: attrs.Updated,
		})
	}
	return files, nil
}
