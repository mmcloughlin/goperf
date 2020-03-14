package suite

import (
	"testing"

	"github.com/google/uuid"
)

func TestFixedUUID(t *testing.T) {
	mod := &Module{
		Path:    "github.com/klauspost/compress",
		Version: "v1.10.3",
	}
	pkg := &Package{
		Module:       mod,
		RelativePath: "huff0",
	}
	benchmark := &Benchmark{
		Package: pkg,
		Name:    "Compress1X",
		Parameters: map[string]string{
			"reuse":  "none",
			"corpus": "pngdata.001",
		},
		Unit: "MB/s",
	}

	cases := []struct {
		Object interface{ UUID() uuid.UUID }
		Expect string
	}{
		{Object: mod, Expect: "c060fae1-5c86-5744-b3f5-3d48dae00294"},
		{Object: pkg, Expect: "8908e73a-5ea4-5953-b3e9-c1259263ba2c"},
		{Object: benchmark, Expect: "e95a5028-cbc3-5b5e-94ac-cf3bab3f1113"},
	}
	for _, c := range cases {
		if got := c.Object.UUID(); got.String() != c.Expect {
			t.Errorf("%#v has uuid %s; expect %s", c.Object, got, c.Expect)
		}
	}
}
