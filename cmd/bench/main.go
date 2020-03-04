package main

import (
	"flag"
	"os"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/lg"
)

func main() {
	logger := lg.Default()
	base := command.NewBase(logger)

	// Platform provides OS specific functionality.
	p := &Platform{}

	// Runner.
	r := NewRun(base, p)
	subcommands.Register(r, "benchmark execution")

	// Wrappers.
	for _, wrapper := range p.Wrappers(base) {
		subcommands.Register(wrapper, "process wrapping")
	}

	// Help.
	subcommands.Register(subcommands.HelpCommand(), "help")
	subcommands.Register(subcommands.CommandsCommand(), "help")

	// Execute.
	flag.Parse()
	ctx := command.BackgroundContext(logger)
	os.Exit(int(subcommands.Execute(ctx)))
}
