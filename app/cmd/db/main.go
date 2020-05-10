package main

import (
	"context"
	"database/sql"
	"flag"

	"github.com/google/subcommands"
	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/pkg/command"

	// Register database drivers
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres" // cloudsqlpostgres
	_ "github.com/lib/pq"                                                    // postgres
)

func main() {
	command.Run(run)
}

// Top-level flags.
var (
	driver = flag.String("driver", "postgres", "database driver")
	conn   = flag.String("conn", "", "database connection string")
)

func run(ctx context.Context, l *zap.Logger) int {
	base := command.NewBase(l)

	// Database commands.
	subcommands.Register(NewMigrate(base), "database admin")
	subcommands.Register(NewTruncate(base), "database admin")

	subcommands.Register(NewCommits(base), "data ingestion")
	subcommands.Register(NewRefs(base), "data ingestion")
	subcommands.Register(NewPositions(base), "data ingestion")
	subcommands.Register(NewAddMod(base), "data ingestion")

	subcommands.Register(NewTraces(base), "data access")
	subcommands.Register(NewChangeTest(base), "data access")

	// Help.
	subcommands.Register(subcommands.HelpCommand(), "help")
	subcommands.Register(subcommands.CommandsCommand(), "help")
	subcommands.Register(subcommands.FlagsCommand(), "help")

	// Execute.
	flag.Parse()
	return int(subcommands.Execute(ctx))
}

// open database connection.
func open() (*sql.DB, error) {
	return sql.Open(*driver, *conn)
}
