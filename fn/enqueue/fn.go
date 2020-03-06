package enqueue

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/mmcloughlin/cb/app/launch"
	"github.com/mmcloughlin/cb/app/repo"
	"github.com/mmcloughlin/cb/pkg/job"
)

// Parameters.
var (
	topic = os.Getenv("CB_JOBS_TOPIC")
)

// FirestoreEvent is the payload of a Firestore event.
type FirestoreEvent struct {
	OldValue   FirestoreValue `json:"oldValue"`
	Value      FirestoreValue `json:"value"`
	UpdateMask struct {
		FieldPaths []string `json:"fieldPaths"`
	} `json:"updateMask"`
}

// FirestoreValue holds Firestore fields.
type FirestoreValue struct {
	CreateTime time.Time         `json:"createTime"`
	Fields     repo.CommitFields `json:"fields"`
	Name       string            `json:"name"`
	UpdateTime time.Time         `json:"updateTime"`
}

// Handle creation of a commit in Firestore.
func Handle(ctx context.Context, e FirestoreEvent) error {
	commit := e.Value.Fields.Commit()
	log.Printf("creation time: %s", e.Value.CreateTime)
	log.Printf("sha: %#v", commit.SHA)

	// Create launcher.
	l, err := launch.NewLauncher(ctx, topic)
	if err != nil {
		return err
	}
	defer l.Close()

	// Publish a job.
	// TODO(mbm): remove hardcoded job
	j := &job.Job{
		Toolchain: job.Toolchain{
			Type: "snapshot",
			Params: map[string]string{
				"builder_type": "linux-amd64",
				"revision":     commit.SHA,
			},
		},
		Suites: []job.Suite{
			{
				Module: job.Module{
					Path:    "github.com/klauspost/compress",
					Version: "b949da471e55fbe4393e6eb595602d936f5c312e",
				},
				Short:     true,
				BenchTime: 10 * time.Millisecond,
			},
		},
	}

	s, err := l.Launch(ctx, j)
	if err != nil {
		return err
	}

	log.Printf("submitted job with id = %s", s.ID)

	return nil
}
