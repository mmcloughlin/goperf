package repo

import (
	"context"

	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/mapper"
	"github.com/mmcloughlin/cb/app/model"
	"github.com/mmcloughlin/cb/app/obj"
)

type Store interface {
	FindBySHA(ctx context.Context, sha string) (*entity.Commit, error)
}

type objstore struct {
	store obj.Store
}

func NewObjectStore(s obj.Store) Store {
	return &objstore{
		store: s,
	}
}

func (s *objstore) FindBySHA(ctx context.Context, sha string) (*entity.Commit, error) {
	obj := new(model.Commit)
	obj.SHA = sha
	if err := s.store.Get(ctx, obj, obj); err != nil {
		return nil, err
	}
	return mapper.CommitFromModel(obj), nil
}
