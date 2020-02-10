package repo

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
)

type Store interface {
	Upsert(ctx context.Context, c *Commit) error
}

type fsstore struct {
	client *firestore.Client
	ref    *firestore.CollectionRef
}

func NewFirestoreStore(c *firestore.Client, collection string) Store {
	return &fsstore{
		client: c,
		ref:    c.Collection(collection),
	}
}

func (s *fsstore) Upsert(ctx context.Context, c *Commit) error {
	// Map to Firestore object.
	obj := tofscommit(c)

	// Write to Firestore.
	_, err := s.ref.Doc(obj.SHA).Set(ctx, obj)
	if err != nil {
		return fmt.Errorf("firestore set: %w", err)
	}

	return nil
}

type fscommit struct {
	SHA            string    `firestore:"sha"`
	Tree           string    `firestore:"tree"`
	Parents        []string  `firestore:"parents"`
	AuthorName     string    `firestore:"author_name"`
	AuthorEmail    string    `firestore:"author_email"`
	AuthorTime     time.Time `firestore:"author_time"`
	CommitterName  string    `firestore:"committer_name"`
	CommitterEmail string    `firestore:"committer_email"`
	CommitTime     time.Time `firestore:"commit_time"`
	Message        string    `firestore:"message"`
}

func tofscommit(c *Commit) *fscommit {
	return &fscommit{
		SHA:            c.SHA,
		Tree:           c.Tree,
		Parents:        c.Parents,
		AuthorName:     c.Author.Name,
		AuthorEmail:    c.Author.Email,
		AuthorTime:     c.AuthorTime,
		CommitterName:  c.Committer.Name,
		CommitterEmail: c.Committer.Email,
		CommitTime:     c.CommitTime,
		Message:        c.Message,
	}
}
