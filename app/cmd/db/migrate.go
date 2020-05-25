package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"
	"github.com/pressly/goose"

	"github.com/mmcloughlin/goperf/pkg/command"
)

type Migrate struct {
	command.Base

	dir string
}

func NewMigrate(b command.Base) *Migrate {
	return &Migrate{
		Base: b,
	}
}

func (*Migrate) Name() string { return "migrate" }

func (*Migrate) Synopsis() string {
	return "migrate database schema"
}

func (*Migrate) Usage() string {
	return ""
}

func (cmd *Migrate) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.dir, "dir", ".", "directory with migration files")
}

func (cmd *Migrate) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) (status subcommands.ExitStatus) {
	args := f.Args()
	if len(args) == 0 {
		return cmd.UsageError("missing command")
	}
	command := args[0]
	args = args[1:]

	// Set goose dialect.
	if err := goose.SetDialect("postgres"); err != nil {
		return cmd.Error(err)
	}

	// Open database.
	d, err := open()
	if err != nil {
		return cmd.Error(err)
	}
	defer cmd.CheckClose(&status, d)

	return cmd.Status(goose.Run(command, d, cmd.dir, args...))
}
