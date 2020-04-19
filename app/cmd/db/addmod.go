package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/google/subcommands"
	"github.com/manifoldco/promptui"
	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/mod"
)

type AddMod struct {
	command.Base
}

func NewAddMod(b command.Base) *AddMod {
	return &AddMod{
		Base: b,
	}
}

func (*AddMod) Name() string { return "addmod" }

func (*AddMod) Synopsis() string {
	return "add a module to the database"
}

func (*AddMod) Usage() string {
	return ""
}

func (cmd *AddMod) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) (status subcommands.ExitStatus) {
	// Process arguments.
	path := f.Arg(0)
	if path == "" {
		return cmd.UsageError("no module path provided")
	}

	// Fetch latest version.
	mdb := mod.NewOfficialModuleProxy(http.DefaultClient)
	latest, err := mdb.Latest(ctx, path)
	if err != nil {
		return cmd.Error(err)
	}

	cmd.Log.Info("found latest version",
		zap.String("path", path),
		zap.String("version", latest.Version),
		zap.Time("time", latest.Time),
	)

	m := &entity.Module{
		Path:    path,
		Version: latest.Version,
	}

	// Confirm.
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Add module %s", m),
		IsConfirm: true,
	}

	if _, err := prompt.Run(); err != nil {
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
	defer cmd.CheckClose(&status, d)

	// Insert module.
	if err := d.StoreModule(ctx, m); err != nil {
		return cmd.Error(err)
	}

	return subcommands.ExitSuccess
}
