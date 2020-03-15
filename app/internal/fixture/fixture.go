package fixture

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/mmcloughlin/cb/app/repo"
	"github.com/mmcloughlin/cb/app/results"
	"github.com/mmcloughlin/cb/app/suite"
)

// Sample model objects for testing purposes.
var (
	ModuleSHA = "788b7f06fee85b7e1d2aa4a3a86f8dbbbcc771ae"

	RevInfo = &suite.RevInfo{
		Version: "v1.10.3",
		Time:    time.Date(2020, 3, 11, 11, 43, 27, 0, time.UTC), // "2020-03-11T11:43:27Z"
	}

	Module = &suite.Module{
		Path:    "github.com/klauspost/compress",
		Version: RevInfo.Version,
	}

	Package = &suite.Package{
		Module:       Module,
		RelativePath: "huff0",
	}

	Benchmark = &suite.Benchmark{
		Package:  Package,
		FullName: "BenchmarkCompress1X/reuse=none/corpus=pngdata.001",
		Name:     "Compress1X",
		Parameters: map[string]string{
			"reuse":  "none",
			"corpus": "pngdata.001",
		},
		Unit: "MB/s",
	}

	Commit = &repo.Commit{
		SHA:     "d401c427b29f48d5cbc5092e62c20aa8524ce356",
		Tree:    "69f35623bf5c665d687eba295db7bc619a5f9f31",
		Parents: []string{"60b9ae4cf3a0428668748a53f278a80d41fbfc38"},
		Author: repo.Person{
			Name:  "Michael McLoughlin",
			Email: "mmcloughlin@gmail.com",
		},
		AuthorTime: time.Date(2017, 15, 7, 18, 21, 26, 0, time.FixedZone("UTC-6", -6*60*60)), // "Sat Jul 15 18:21:26 2017 -0600"
		Committer: repo.Person{
			Name:  "Adam Langley",
			Email: "agl@golang.org",
		},
		CommitTime: time.Date(2017, 8, 9, 19, 29, 14, 0, time.UTC), // "Wed Aug 09 19:29:14 2017 +0000"
		Message:    "crypto/rand: batch large calls to linux getrandom",
	}

	DataFile = &results.DataFile{
		Name:   "e5a4b8be-c0e5-42c4-a243-2b458ceff483.txt",
		SHA256: decodesha256("a36064ebcadaea9b4b419ef66b487a6bdf1f0d5f90efa513d35f800d4dfceeb1"),
	}

	Result = &results.Result{
		File:      DataFile,
		Line:      42,
		Benchmark: Benchmark,
		Commit:    Commit,
		Environment: results.Properties{
			"a": "1",
			"b": "2",
		},
		Metadata: results.Properties{
			"c": "3",
			"d": "4",
		},
		Iterations: 4096,
		Value:      123.45,
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
