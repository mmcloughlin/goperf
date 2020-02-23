package wrap

import (
	"os"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/pkg/runner"
)

// RunUnder builds a wrapper that runs under the given subcommand, assuming that
// subcommand is registered on this executable.
func RunUnder(cmd subcommands.Command) (runner.Wrapper, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	return runner.RunUnder(self, cmd.Name()), nil
}
