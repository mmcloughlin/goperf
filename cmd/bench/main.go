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

	// Runner.
	r := NewRun(base, p)
	subcommands.Register(r, "benchmark execution")

	// Wrappers.
	for _, wrapper := range p.Wrappers() {
		subcommands.Register(wrapper, "process wrapping")
	}

	// Help.
	subcommands.Register(subcommands.HelpCommand(), "help")
	subcommands.Register(subcommands.CommandsCommand(), "help")

	// Execute.
	flag.Parse()
	return int(subcommands.Execute(ctx))
}
