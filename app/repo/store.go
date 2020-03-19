package repo

import (
	"context"

	"github.com/mmcloughlin/cb/app/model"
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
	obj := new(model.Commit)
	obj.SHA = sha
	if err := s.store.Get(ctx, obj, obj); err != nil {
		return nil, err
	}
	return frommodelcommit(obj), nil
}

func (s *objstore) Upsert(ctx context.Context, c *Commit) error {
	// Map to object.
	obj := tomodelcommit(c)

	// Write to Firestore.
	return s.store.Set(ctx, obj)
}

func tomodelcommit(c *Commit) *model.Commit {
	return &model.Commit{
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

func frommodelcommit(c *model.Commit) *Commit {
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
