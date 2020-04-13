// Package fixture provides fake objects for testing.
package fixture

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/pkg/mod"
)

// Sample entity objects for testing purposes.
var (
	ModuleSHA = "788b7f06fee85b7e1d2aa4a3a86f8dbbbcc771ae"

	RevInfo = &mod.RevInfo{
		Version: "v1.10.3",
		Time:    time.Date(2020, 3, 11, 11, 43, 27, 0, time.UTC), // "2020-03-11T11:43:27Z"
	}

	Module = &entity.Module{
		Path:    "github.com/klauspost/compress",
		Version: RevInfo.Version,
	}

	Package = &entity.Package{
		Module:       Module,
		RelativePath: "huff0",
	}

	Benchmark = &entity.Benchmark{
		Package:  Package,
		FullName: "BenchmarkCompress1X/reuse=none/corpus=pngdata.001",
		Name:     "Compress1X",
		Parameters: map[string]string{
			"reuse":  "none",
			"corpus": "pngdata.001",
		},
		Unit: "MB/s",
	}

	Commit = &entity.Commit{
		SHA:     "d401c427b29f48d5cbc5092e62c20aa8524ce356",
		Tree:    "69f35623bf5c665d687eba295db7bc619a5f9f31",
		Parents: []string{"60b9ae4cf3a0428668748a53f278a80d41fbfc38"},
		Author: entity.Person{
			Name:  "Michael McLoughlin",
			Email: "mmcloughlin@gmail.com",
		},
		AuthorTime: time.Date(2017, 15, 7, 18, 21, 26, 0, time.FixedZone("UTC-6", -6*60*60)), // "Sat Jul 15 18:21:26 2017 -0600"
		Committer: entity.Person{
			Name:  "Adam Langley",
			Email: "agl@golang.org",
		},
		CommitTime: time.Date(2017, 8, 9, 19, 29, 14, 0, time.UTC), // "Wed Aug 09 19:29:14 2017 +0000"
		Message:    "crypto/rand: batch large calls to linux getrandom",
	}

	DataFile = &entity.DataFile{
		Name:   "e5a4b8be-c0e5-42c4-a243-2b458ceff483.txt",
		SHA256: decodesha256("a36064ebcadaea9b4b419ef66b487a6bdf1f0d5f90efa513d35f800d4dfceeb1"),
	}

	Result = &entity.Result{
		File:      DataFile,
		Line:      42,
		Benchmark: Benchmark,
		Commit:    Commit,
		Environment: entity.Properties{
			"a": "1",
			"b": "2",
		},
		Metadata: entity.Properties{
			"c": "3",
			"d": "4",
		},
		Iterations: 4096,
		Value:      123.45,
	}

	Worker = "gopher"

	TaskSpec = entity.TaskSpec{
		Type:       entity.TaskTypeModule,
		TargetUUID: Module.UUID(),
		CommitSHA:  Commit.SHA,
	}

	Task = &entity.Task{
		UUID:             uuid.MustParse("6e68e764-e3fd-4e86-b122-e07382dd57b0"),
		Worker:           Worker,
		Spec:             TaskSpec,
		Status:           entity.TaskStatusCreated,
		LastStatusUpdate: time.Date(2020, 4, 7, 20, 50, 13, 0, time.UTC),
		DatafileUUID:     uuid.Nil,
	}
)

func decodesha256(s string) [sha256.Size]byte {
	if len(s)*4 != 256 {
		panic("too short")
	}
	sha, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	var out [sha256.Size]byte
	copy(out[:], sha)
	return out
}
