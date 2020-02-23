package main

import (
	"flag"
	"log"
	"os"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/lg"
	"github.com/mmcloughlin/cb/pkg/wrap"
)

func main() {
	logger := lg.Default()
	base := command.NewBase(logger)

	// Runner.
	r := NewRun(base)
	subcommands.Register(r, "benchmark execution")

	// Wrappers.
	wrapcmds := []subcommands.Command{
		wrap.NewConfigDefault(base),
		wrap.NewPrioritize(base),
	}
	for _, wrapcmd := range wrapcmds {
		subcommands.Register(wrapcmd, "process wrapping")
		w, err := wrap.RunUnder(wrapcmd)
		if err != nil {
			log.Fatal(err)
		}
		r.AddWrapper(w)
	}

	// Help.
	subcommands.Register(subcommands.HelpCommand(), "help")
	subcommands.Register(subcommands.CommandsCommand(), "help")

	// Execute.
	flag.Parse()
	ctx := command.BackgroundContext(logger)
	os.Exit(int(subcommands.Execute(ctx)))
}
