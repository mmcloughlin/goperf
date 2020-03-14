package fixture

import (
	"testing"

	"github.com/google/uuid"
)

func TestFixedUUID(t *testing.T) {
	cases := []struct {
		Object interface{ UUID() uuid.UUID }
		Expect string
	}{
		{Object: Module, Expect: "c060fae1-5c86-5744-b3f5-3d48dae00294"},
		{Object: Package, Expect: "8908e73a-5ea4-5953-b3e9-c1259263ba2c"},
		{Object: Benchmark, Expect: "e95a5028-cbc3-5b5e-94ac-cf3bab3f1113"},
		{Object: DataFile, Expect: "fc270fc5-5b11-54d1-93e7-43e3431aeb7a"},
		{Object: Result, Expect: "5a4da14d-29da-50b9-bc51-db5e67a8ef1b"},
	}
	for _, c := range cases {
		if got := c.Object.UUID(); got.String() != c.Expect {
			t.Errorf("%#v has uuid %s; expect %s", c.Object, got, c.Expect)
		}
	}
}
