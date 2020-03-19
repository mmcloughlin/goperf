package repo

import (
	"context"
	"time"

	"github.com/mmcloughlin/cb/app/obj"
)

type Store interface {
	FindBySHA(ctx context.Context, sha string) (*Commit, error)
	Upsert(ctx context.Context, c *Commit) error
}

type objstore struct {
	store obj.Store
}

func NewObjectStore(s obj.Store) Store {
	return &objstore{
		store: s,
	}
}

func (s *objstore) FindBySHA(ctx context.Context, sha string) (*Commit, error) {
	obj := new(fscommit)
	obj.SHA = sha
	if err := s.store.Get(ctx, obj, obj); err != nil {
		return nil, err
	}
	return obj.Commit(), nil
}

func (s *objstore) Upsert(ctx context.Context, c *Commit) error {
	// Map to object.
	obj := tofscommit(c)

	// Write to Firestore.
	return s.store.Set(ctx, obj)
}

type fscommit struct {
	SHA            string    `firestore:"sha" json:"sha"`
	Tree           string    `firestore:"tree" json:"tree"`
	Parents        []string  `firestore:"parents" json:"parents"`
	AuthorName     string    `firestore:"author_name" json:"author_name"`
	AuthorEmail    string    `firestore:"author_email" json:"author_email"`
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

func (c *fscommit) Type() string { return "commits" }
func (c *fscommit) ID() string   { return c.SHA }

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
