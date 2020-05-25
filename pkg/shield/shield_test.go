package shield

import (
	"testing"

	"github.com/mmcloughlin/goperf/pkg/runner"
)

func TestShieldImplementsTuner(t *testing.T) {
	var _ runner.Tuner = new(Shield)
}
