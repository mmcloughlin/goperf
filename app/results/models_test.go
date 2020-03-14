package results

import (
	"encoding/hex"
	"testing"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/suite"
)

func TestFixedUUID(t *testing.T) {
	mod := &suite.Module{
		Path:    "github.com/klauspost/compress",
		Version: "v1.10.3",
	}

	pkg := &suite.Package{
		Module:       mod,
		RelativePath: "huff0",
	}

	benchmark := &suite.Benchmark{
		Package: pkg,
		Name:    "Compress1X",
		Parameters: map[string]string{
			"reuse":  "none",
			"corpus": "pngdata.001",
		},
		Unit: "MB/s",
	}

	df := &DataFile{
		Name: "e5a4b8be-c0e5-42c4-a243-2b458ceff483.txt",
	}
	sha, err := hex.DecodeString("a36064ebcadaea9b4b419ef66b487a6bdf1f0d5f90efa513d35f800d4dfceeb1")
	if err != nil {
		t.Fatal(err)
	}
	copy(df.SHA256[:], sha)

	r := &Result{
		File:      df,
		Line:      42,
		Benchmark: benchmark,
		Commit:    nil,
		Environment: Properties{
			"a": "1",
			"b": "2",
		},
		Metadata: Properties{
			"c": "3",
			"d": "4",
		},
		Value: 123.45,
	}

	cases := []struct {
		Object interface{ UUID() uuid.UUID }
		Expect string
	}{
		{Object: df, Expect: "fc270fc5-5b11-54d1-93e7-43e3431aeb7a"},
		{Object: r, Expect: "5a4da14d-29da-50b9-bc51-db5e67a8ef1b"},
	}
	for _, c := range cases {
		if got := c.Object.UUID(); got.String() != c.Expect {
			t.Errorf("%#v has uuid %s; expect %s", c.Object, got, c.Expect)
		}
	}
}
