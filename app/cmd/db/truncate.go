package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"
	"github.com/manifoldco/promptui"

	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/pkg/command"
)

type Truncate struct {
	command.Base
}

func NewTruncate(b command.Base) *Truncate {
	return &Truncate{
		Base: b,
	}
}

func (*Truncate) Name() string { return "truncate" }

func (*Truncate) Synopsis() string {
	return "delete data from the database"
}

func (*Truncate) Usage() string {
	return ""
}

func (cmd *Truncate) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// Destructive action. Ask for confirmation first.
	prompt := promptui.Prompt{
		Label:     "This is a destructive action. Are you sure you want to delete all data in the database",
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		return subcommands.ExitSuccess
	}

	// Open database.
	sqldb, err := open()
	if err != nil {
		return cmd.Error(err)
	}

	d, err := db.New(ctx, sqldb)
	if err != nil {
		return cmd.Error(err)
	}
	defer d.Close()

	// Call truncate.
	if err := d.TruncateNonStatic(ctx); err != nil {
		return cmd.Error(err)
	}

	cmd.Log.Info("data deleted")

	return subcommands.ExitSuccess
}
