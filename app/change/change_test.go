package change

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		Pre, Post float64
		Unit      string
		Expect    Type
	}{
		{1, 2, "MB/s", TypeImprovement},
		{1, 0.5, "MB/s", TypeRegression},
		{1, 1, "MB/s", TypeUnchanged},

		{1, 2, "ns/op", TypeRegression},
		{1, 0.5, "ns/op", TypeImprovement},
		{1, 1, "ns/op", TypeUnchanged},

		{1, 2, "frobs/s", TypeUnknown},
		{1, 0.5, "frobs/s", TypeUnknown},
		{1, 1, "frobs/s", TypeUnchanged},
	}
	for _, c := range cases {
		if got := Classify(c.Pre, c.Post, c.Unit); got != c.Expect {
			t.Errorf("Classify(%v, %v, %q) = %v; expect %v", c.Pre, c.Post, c.Unit, got, c.Expect)
		}
	}
}
