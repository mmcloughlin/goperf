package shield

import (
	"testing"

	"github.com/mmcloughlin/cb/pkg/cpuset"
	"github.com/mmcloughlin/cb/pkg/runner"
)

func TestShieldImplementsTuner(t *testing.T) {
	var _ runner.Tuner = new(Shield)
}

func TestPick(t *testing.T) {
	s := cpuset.NewSet(1, 2, 3, 4, 5, 6, 7, 8)
	sub, err := pick(s, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(sub) != 5 {
		t.Fatal("wrong size")
	}
	if !s.Contains(sub) {
		t.Fatal("not subset")
	}
}
