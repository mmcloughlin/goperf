package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"
	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/pkg/command"
	"github.com/mmcloughlin/goperf/pkg/platform"
)

func main() {
	command.Run(run)
}

func run(ctx context.Context, l *zap.Logger) int {
	base := command.NewBase(l)

	// Platform provides OS specific functionality.
	p := platform.New(base)

	// Run worker loop command.
	r := NewRun(base, p)
	subcommands.Register(r, "job processing")

	// Wrappers.
	for _, wrapper := range p.Wrappers() {
		subcommands.Register(wrapper, "internal use")
	}

	// Help.
	subcommands.Register(subcommands.HelpCommand(), "help")
	subcommands.Register(subcommands.CommandsCommand(), "help")

	// Execute.
	flag.Parse()
	return int(subcommands.Execute(ctx))
}
