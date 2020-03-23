package service

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"

	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/mapper"
	"github.com/mmcloughlin/cb/app/model"
	"github.com/mmcloughlin/cb/app/obj"
)

type Commits interface {
	FindCommitBySHA(ctx context.Context, sha string) (*entity.Commit, error)
}

type Packages interface {
	FindModuleByUUID(ctx context.Context, id uuid.UUID) (*entity.Module, error)
	ListModules(ctx context.Context) ([]*entity.Module, error)

	FindPackageByUUID(ctx context.Context, id uuid.UUID) (*entity.Package, error)
	ListModulePackages(ctx context.Context, m *entity.Module) ([]*entity.Package, error)
}

type Benchmarks interface {
	FindBenchmarkByUUID(ctx context.Context, id uuid.UUID) (*entity.Benchmark, error)
	ListPackageBenchmarks(ctx context.Context, p *entity.Package) ([]*entity.Benchmark, error)
}

type Results interface {
	ListBenchmarkResults(ctx context.Context, b *entity.Benchmark) ([]*entity.Result, error)

	FindDataFileByUUID(ctx context.Context, id uuid.UUID) (*entity.DataFile, error)

	FindPropertiesByUUID(ctx context.Context, id uuid.UUID) (entity.Properties, error)
}

type Service interface {
	Commits
	Packages
	Benchmarks
	Results
}

type fire struct {
	client *firestore.Client
	store  obj.Store
}

func NewFirestore(c *firestore.Client, caches ...obj.Store) Service {
	stores := append(caches, obj.NewFirestore(c))
	return &fire{
		client: c,
		store:  obj.Overlay(stores...),
	}
}

func (f *fire) FindCommitBySHA(ctx context.Context, sha string) (*entity.Commit, error) {
	obj := new(model.Commit)
	obj.SHA = sha
	if err := f.store.Get(ctx, obj, obj); err != nil {
		return nil, err
	}
	return mapper.CommitFromModel(obj), nil
}

func (f *fire) FindModuleByUUID(ctx context.Context, id uuid.UUID) (*entity.Module, error) {
	obj := new(model.Module)
	obj.UUID = id.String()
	if err := f.store.Get(ctx, obj, obj); err != nil {
		return nil, err
	}
	return mapper.ModuleFromModel(obj), nil
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

func (f *fire) FindPackageByUUID(ctx context.Context, id uuid.UUID) (*entity.Package, error) {
	// Get the package.
	obj := new(model.Package)
	obj.UUID = id.String()
	if err := f.store.Get(ctx, obj, obj); err != nil {
		return nil, err
	}

	// Get the associated module.
	modid, err := uuid.Parse(obj.ModuleUUID)
	if err != nil {
		return nil, err
	}

	mod, err := f.FindModuleByUUID(ctx, modid)
	if err != nil {
		return nil, err
	}

	return mapper.PackageFromModel(obj, mod), nil
}

func (f *fire) ListModulePackages(ctx context.Context, m *entity.Module) ([]*entity.Package, error) {
	iter := f.Collection(&model.Package{}).Where("module_uuid", "==", m.UUID().String()).Documents(ctx)

	var pkgs []*entity.Package
	for {
		docsnap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		pkg := &model.Package{}
		if err := docsnap.DataTo(pkg); err != nil {
			return nil, err
		}

		pkgs = append(pkgs, mapper.PackageFromModel(pkg, m))
	}

	return pkgs, nil
}

func (f *fire) FindBenchmarkByUUID(ctx context.Context, id uuid.UUID) (*entity.Benchmark, error) {
	// Get the benchmark.
	obj := new(model.Benchmark)
	obj.UUID = id.String()
	if err := f.store.Get(ctx, obj, obj); err != nil {
		return nil, err
	}

	// Get the associated package.
	pkgid, err := uuid.Parse(obj.PackageUUID)
	if err != nil {
		return nil, err
	}

	pkg, err := f.FindPackageByUUID(ctx, pkgid)
	if err != nil {
		return nil, err
	}

	return mapper.BenchmarkFromModel(obj, pkg), nil
}

func (f *fire) ListPackageBenchmarks(ctx context.Context, p *entity.Package) ([]*entity.Benchmark, error) {
	iter := f.Collection(&model.Benchmark{}).Where("package_uuid", "==", p.UUID().String()).Documents(ctx)

	var benchs []*entity.Benchmark
	for {
		docsnap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		bench := &model.Benchmark{}
		if err := docsnap.DataTo(bench); err != nil {
			return nil, err
		}

		benchs = append(benchs, mapper.BenchmarkFromModel(bench, p))
	}

	return benchs, nil
}

func (f *fire) ListBenchmarkResults(ctx context.Context, b *entity.Benchmark) ([]*entity.Result, error) {
	iter := f.Collection(&model.Result{}).Where("benchmark_uuid", "==", b.UUID().String()).Documents(ctx)

	var results []*entity.Result
	for {
		docsnap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		resultmodel := &model.Result{}
		if err := docsnap.DataTo(resultmodel); err != nil {
			return nil, err
		}

		result, err := f.result(ctx, resultmodel)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

func (f *fire) FindDataFileByUUID(ctx context.Context, id uuid.UUID) (*entity.DataFile, error) {
	obj := new(model.DataFile)
	obj.UUID = id.String()
	if err := f.store.Get(ctx, obj, obj); err != nil {
		return nil, err
	}
	return mapper.DataFileFromModel(obj), nil
}

func (f *fire) FindPropertiesByUUID(ctx context.Context, id uuid.UUID) (entity.Properties, error) {
	obj := new(model.Properties)
	obj.UUID = id.String()
	if err := f.store.Get(ctx, obj, obj); err != nil {
		return nil, err
	}
	return mapper.PropertiesFromModel(obj), nil
}

func (f *fire) result(ctx context.Context, r *model.Result) (*entity.Result, error) {
	// DataFile.
	fileid, err := uuid.Parse(r.DataFileUUID)
	if err != nil {
		return nil, err
	}
	file, err := f.FindDataFileByUUID(ctx, fileid)
	if err != nil {
		return nil, err
	}

	// Benchmark.
	benchid, err := uuid.Parse(r.BenchmarkUUID)
	if err != nil {
		return nil, err
	}
	bench, err := f.FindBenchmarkByUUID(ctx, benchid)
	if err != nil {
		return nil, err
	}

	// Commit.
	commit, err := f.FindCommitBySHA(ctx, r.CommitSHA)
	if err != nil {
		return nil, err
	}

	// Environment.
	envid, err := uuid.Parse(r.EnvironmentUUID)
	if err != nil {
		return nil, err
	}
	env, err := f.FindPropertiesByUUID(ctx, envid)
	if err != nil {
		return nil, err
	}

	// Metadata.
	metaid, err := uuid.Parse(r.MetadataUUID)
	if err != nil {
		return nil, err
	}
	meta, err := f.FindPropertiesByUUID(ctx, metaid)
	if err != nil {
		return nil, err
	}

	return mapper.ResultsFromModel(
		r,
		file,
		bench,
		commit,
		env,
		meta,
	), nil
}

func (f *fire) Collection(k obj.Key) *firestore.CollectionRef {
	return f.client.Collection(k.Type())
}
