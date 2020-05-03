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

	meta bool
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

func (cmd *AddMod) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.meta, "meta", false, "special-case meta module such as \"std\" or \"cmd\"")
}

func (cmd *AddMod) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) (status subcommands.ExitStatus) {
	// Process arguments.
	path := f.Arg(0)
	if path == "" {
		return cmd.UsageError("no module path provided")
	}

	// Fetch latest version.
	v, err := cmd.version(ctx, path)
	if err != nil {
		return cmd.Error(err)
	}

	m := &entity.Module{
		Path:    path,
		Version: v,
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

func (cmd *AddMod) version(ctx context.Context, path string) (string, error) {
	if cmd.meta {
		cmd.Log.Info("meta module assumed versionless")
		return "", nil
	}

	mdb := mod.NewOfficialModuleProxy(http.DefaultClient)
	latest, err := mdb.Latest(ctx, path)
	if err != nil {
		return "", err
	}

	cmd.Log.Info("found latest version",
		zap.String("path", path),
		zap.String("version", latest.Version),
		zap.Time("time", latest.Time),
	)

	return latest.Version, nil
}
