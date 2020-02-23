package main

import (
	"flag"
	"os"

	"github.com/mmcloughlin/cb/pkg/wrap"

	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/lg"

	"github.com/google/subcommands"
)

func main() {
	logger := lg.Default()
	ctx := command.BackgroundContext(logger)

	base := command.NewBase(logger)
	subcommands.Register(NewRun(base), "")
	subcommands.Register(wrap.NewConfigDefault(base), "")
	subcommands.Register(subcommands.HelpCommand(), "")

	flag.Parse()
	os.Exit(int(subcommands.Execute(ctx)))
}
