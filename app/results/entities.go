package results

import (
	"crypto/sha256"
	"strconv"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/id"
	"github.com/mmcloughlin/cb/app/repo"
	"github.com/mmcloughlin/cb/app/suite"
)

type DataFile struct {
	Name   string
	SHA256 [sha256.Size]byte
}

var datafilenamespace = uuid.MustParse("3e094884-6ffd-4d70-a83f-bc2d241b7344")

func (f *DataFile) UUID() uuid.UUID {
	return id.UUID(datafilenamespace, f.SHA256[:])
}

type Result struct {
	File        *DataFile
	Line        int
	Benchmark   *suite.Benchmark
	Commit      *repo.Commit
	Environment Properties
	Metadata    Properties
	Iterations  uint64
	Value       float64
}

var resultnamespace = uuid.MustParse("0063a4c4-2bdc-4c3b-878b-5c90356013a3")

func (r *Result) UUID() uuid.UUID {
	return id.Strings(resultnamespace, []string{
		r.File.UUID().String(),
		strconv.Itoa(r.Line),
		r.Benchmark.UUID().String(),
	})
}

type Properties map[string]string

var propertiesnamespace = uuid.MustParse("d0c136af-cf22-4f7a-87b3-4a73bfb57489")

func (p Properties) UUID() uuid.UUID {
	return id.KeyValues(propertiesnamespace, p)
}
