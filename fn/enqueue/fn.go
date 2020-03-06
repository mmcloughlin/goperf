package enqueue

import (
	"context"
	"log"
	"time"

	"github.com/mmcloughlin/cb/app/repo"
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
	return nil
}
