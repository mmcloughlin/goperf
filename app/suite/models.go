package suite

import (
	"path"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/id"
)

type Module struct {
	Path    string
	Version string
}

var modulenamespace = uuid.MustParse("24662676-cba4-4241-ab2d-e81de0d407b4")

func (m *Module) UUID() uuid.UUID {
	return id.Strings(modulenamespace, []string{m.Path, m.Version})
}

type Package struct {
	Module       *Module
	RelativePath string
}

var packagenamespace = uuid.MustParse("91e2ea8d-5830-4b70-b26c-68ad426636eb")

func (p *Package) UUID() uuid.UUID {
	return id.Strings(packagenamespace, []string{
		p.Module.UUID().String(),
		p.RelativePath,
	})
}

func (p *Package) ImportPath() string {
	return path.Join(p.Module.Path, p.RelativePath)
}

type Benchmark struct {
	Package    *Package
	FullName   string
	Name       string
	Parameters map[string]string
	Unit       string
}

var benchmarknamespace = uuid.MustParse("51d0f236-a868-48f4-8ef1-8d303e4953e1")

func (b *Benchmark) UUID() uuid.UUID {
	paramsid := id.KeyValues(uuid.Nil, b.Parameters)
	return id.Strings(benchmarknamespace, []string{
		b.Package.UUID().String(),
		b.Name,
		paramsid.String(),
		b.Unit,
	})
}
