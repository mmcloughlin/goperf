package suite

import (
	"context"

	"github.com/mmcloughlin/cb/app/model"
	"github.com/mmcloughlin/cb/app/obj"
)

// Store provides storage for benchmark suite objects.
type Store interface {
	UpsertModule(ctx context.Context, m *Module) error
	UpsertPackage(ctx context.Context, p *Package) error
	UpsertBenchmark(ctx context.Context, b *Benchmark) error
}

type objstore struct {
	store obj.Store
}

// NewObjectStore provides benchmark suite storage backed by an object store.
func NewObjectStore(s obj.Store) Store {
	return &objstore{
		store: s,
	}
}

func (s *objstore) UpsertModule(ctx context.Context, m *Module) error {
	return s.store.Set(ctx, tomodelmodule(m))
}

func (s *objstore) UpsertPackage(ctx context.Context, p *Package) error {
	return s.store.Set(ctx, tomodelpackage(p))
}

func (s *objstore) UpsertBenchmark(ctx context.Context, b *Benchmark) error {
	return s.store.Set(ctx, tomodelbenchmark(b))
}

func tomodelmodule(m *Module) *model.Module {
	return &model.Module{
		UUID:    m.UUID().String(),
		Path:    m.Path,
		Version: m.Version,
	}
}

func tomodelpackage(p *Package) *model.Package {
	return &model.Package{
		UUID:         p.UUID().String(),
		ModuleUUID:   p.Module.UUID().String(),
		RelativePath: p.RelativePath,
	}
}

func tomodelbenchmark(b *Benchmark) *model.Benchmark {
	return &model.Benchmark{
		UUID:        b.UUID().String(),
		PackageUUID: b.Package.UUID().String(),
		Name:        b.Name,
		Unit:        b.Unit,
		Parameters:  b.Parameters,
	}
}
