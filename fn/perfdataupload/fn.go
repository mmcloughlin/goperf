package perfdataupload

import (
	"context"
	"io"
	"log"

	"cloud.google.com/go/storage"
	"golang.org/x/build/buildenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	perfstorage "golang.org/x/perf/storage"
	oauth2api "google.golang.org/api/oauth2/v2"
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

	// GCS client.
	gcs, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	// Perf storage client.
	perfstore, err := NewPerfStorageClient(ctx)
	if err != nil {
		return err
	}

	// Open the GCS object.
	obj := gcs.Bucket(e.Bucket).Object(e.Name)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return err
	}
	defer r.Close()

	// Start an upload.
	u := perfstore.NewUpload(ctx)
	w, err := u.CreateFile(e.Name)
	if err != nil {
		u.Abort()
		return err
	}

	// Transfer from bucket to perf store.
	if _, err := io.Copy(w, r); err != nil {
		u.Abort()
		return err
	}

	// Finalize.
	status, err := u.Commit()
	if err != nil {
		return err
	}

	log.Printf("upload id: %v\n", status.UploadID)
	log.Printf("view url: %v\n", status.ViewURL)

	return nil
}

func NewPerfStorageClient(ctx context.Context) (*perfstorage.Client, error) {
	creds, err := google.FindDefaultCredentials(ctx, oauth2api.UserinfoEmailScope)
	if err != nil {
		return nil, err
	}

	return &perfstorage.Client{
		BaseURL:    buildenv.Production.PerfDataURL,
		HTTPClient: oauth2.NewClient(ctx, creds.TokenSource),
	}, nil
}
