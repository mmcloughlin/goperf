// Package mapper maps internal types to models.
package mapper

import (
	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/model"
	"github.com/mmcloughlin/cb/app/obj"
)

func CommitModel(c *entity.Commit) *model.Commit {
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

func CommitFromModel(c *model.Commit) *entity.Commit {
	return &entity.Commit{
		SHA:     c.SHA,
		Tree:    c.Tree,
		Parents: c.Parents,
		Author: entity.Person{
			Name:  c.AuthorName,
			Email: c.AuthorEmail,
		},
		AuthorTime: c.AuthorTime,
		Committer: entity.Person{
			Name:  c.CommitterName,
			Email: c.CommitterEmail,
		},
		CommitTime: c.CommitTime,
		Message:    c.Message,
	}
}

func ModuleModel(m *entity.Module) *model.Module {
	return &model.Module{
		UUID:    m.UUID().String(),
		Path:    m.Path,
		Version: m.Version,
	}
}

func ModuleFromModel(m *model.Module) *entity.Module {
	return &entity.Module{
		Path:    m.Path,
		Version: m.Version,
	}
}

func PackageModel(p *entity.Package) *model.Package {
	return &model.Package{
		UUID:         p.UUID().String(),
		ModuleUUID:   p.Module.UUID().String(),
		RelativePath: p.RelativePath,
	}
}

func PackageFromModel(p *model.Package, m *entity.Module) *entity.Package {
	return &entity.Package{
		Module:       m,
		RelativePath: p.RelativePath,
	}
}

func PackageModels(p *entity.Package) []obj.Object {
	return []obj.Object{
		ModuleModel(p.Module),
		PackageModel(p),
	}
}

func BenchmarkModel(b *entity.Benchmark) *model.Benchmark {
	return &model.Benchmark{
		UUID:        b.UUID().String(),
		PackageUUID: b.Package.UUID().String(),
		FullName:    b.FullName,
		Name:        b.Name,
		Unit:        b.Unit,
		Parameters:  b.Parameters,
	}
}

func BenchmarkModels(b *entity.Benchmark) []obj.Object {
	return append(
		PackageModels(b.Package),
		BenchmarkModel(b),
	)
}

func DataFileModel(f *entity.DataFile) *model.DataFile {
	return &model.DataFile{
		UUID:   f.UUID().String(),
		Name:   f.Name,
		SHA256: f.SHA256[:],
	}
}

func PropertiesModel(f entity.Properties) *model.Properties {
	return &model.Properties{
		UUID:   f.UUID().String(),
		Fields: f,
	}
}

func ResultModel(r *entity.Result) *model.Result {
	return &model.Result{
		UUID:            r.UUID().String(),
		DataFileUUID:    r.File.UUID().String(),
		Line:            r.Line,
		BenchmarkUUID:   r.Benchmark.UUID().String(),
		CommitSHA:       r.Commit.SHA,
		EnvironmentUUID: r.Environment.UUID().String(),
		MetadataUUID:    r.Metadata.UUID().String(),
		Iterations:      int64(r.Iterations), // model must be int64 since firestore does not support uint64
		Value:           r.Value,
	}
}

func ResultsModels(r *entity.Result) []obj.Object {
	return append(
		BenchmarkModels(r.Benchmark),
		DataFileModel(r.File),
		BenchmarkModel(r.Benchmark),
		CommitModel(r.Commit),
		PropertiesModel(r.Environment),
		PropertiesModel(r.Metadata),
		ResultModel(r),
	)
}
