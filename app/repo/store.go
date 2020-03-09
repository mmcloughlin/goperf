package repo

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
)

type Store interface {
	FindBySHA(ctx context.Context, sha string) (*Commit, error)
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

func (s *fsstore) FindBySHA(ctx context.Context, sha string) (*Commit, error) {
	// Fetch document from Firestore.
	docsnap, err := s.ref.Doc(sha).Get(ctx)
	if err != nil {
		return nil, err
	}

	// Unmarshal.
	var obj fscommit
	if err := docsnap.DataTo(&obj); err != nil {
		return nil, err
	}

	return obj.Commit(), nil
}

func (s *fsstore) Upsert(ctx context.Context, c *Commit) error {
	// Map to Firestore object.
	obj := tofscommit(c)

	// Write to Firestore.
	_, err := s.ref.Doc(obj.SHA).Set(ctx, obj)
	if err != nil {
		return err
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

func (c *fscommit) Commit() *Commit {
	return &Commit{
		SHA:     c.SHA,
		Tree:    c.Tree,
		Parents: c.Parents,
		Author: Person{
			Name:  c.AuthorName,
			Email: c.AuthorEmail,
		},
		AuthorTime: c.AuthorTime,
		Committer: Person{
			Name:  c.CommitterName,
			Email: c.CommitterEmail,
		},
		CommitTime: c.CommitTime,
		Message:    c.Message,
	}
}
