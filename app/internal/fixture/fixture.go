package fixture

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/mmcloughlin/cb/app/results"
	"github.com/mmcloughlin/cb/app/suite"
)

// Sample model objects for testing purposes.
var (
	Module = &suite.Module{
		Path:    "github.com/klauspost/compress",
		Version: "v1.10.3",
	}

	Package = &suite.Package{
		Module:       Module,
		RelativePath: "huff0",
	}

	Benchmark = &suite.Benchmark{
		Package: Package,
		Name:    "Compress1X",
		Parameters: map[string]string{
			"reuse":  "none",
			"corpus": "pngdata.001",
		},
		Unit: "MB/s",
	}

	DataFile = &results.DataFile{
		Name:   "e5a4b8be-c0e5-42c4-a243-2b458ceff483.txt",
		SHA256: decodesha256("a36064ebcadaea9b4b419ef66b487a6bdf1f0d5f90efa513d35f800d4dfceeb1"),
	}

	Result = &results.Result{
		File:      DataFile,
		Line:      42,
		Benchmark: Benchmark,
		Commit:    nil,
		Environment: results.Properties{
			"a": "1",
			"b": "2",
		},
		Metadata: results.Properties{
			"c": "3",
			"d": "4",
		},
		Value: 123.45,
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
