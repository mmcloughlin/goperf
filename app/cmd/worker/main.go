package main

import (
	"flag"
	"os"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/lg"
	"github.com/mmcloughlin/cb/pkg/platform"
)

func main() {
	logger := lg.Default()
	base := command.NewBase(logger)

	// Platform provides OS specific functionality.
	p := &platform.Platform{}

	// Run worker loop command.
	r := NewRun(base, p)
	subcommands.Register(r, "job processing")

	// Wrappers.
	for _, wrapper := range p.Wrappers(base) {
		subcommands.Register(wrapper, "internal use")
	}

	// Help.
	subcommands.Register(subcommands.HelpCommand(), "help")
	subcommands.Register(subcommands.CommandsCommand(), "help")

	// Execute.
	flag.Parse()
	ctx := command.BackgroundContext(logger)
	os.Exit(int(subcommands.Execute(ctx)))
}
