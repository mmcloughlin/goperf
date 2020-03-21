package service

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/mapper"
	"github.com/mmcloughlin/cb/app/model"
	"github.com/mmcloughlin/cb/app/obj"
)

type Service interface {
	ListModules(ctx context.Context) ([]*entity.Module, error)
}

type fire struct {
	client *firestore.Client
}

func NewFirestore(c *firestore.Client) Service {
	return &fire{
		client: c,
	}
}

func (f *fire) ListModules(ctx context.Context) ([]*entity.Module, error) {
	iter := f.Collection(&model.Module{}).Documents(ctx)
	defer iter.Stop()

	var mods []*entity.Module
	for {
		docsnap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		mod := &model.Module{}
		if err := docsnap.DataTo(mod); err != nil {
			return nil, err
		}

		mods = append(mods, mapper.ModuleFromModel(mod))
	}

	return mods, nil
}

func (f *fire) Collection(k obj.Key) *firestore.CollectionRef {
	return f.client.Collection(k.Type())
}
