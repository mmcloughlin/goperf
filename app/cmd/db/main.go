package main

import (
	"database/sql"
	"flag"
	"os"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/lg"

	// Register database drivers
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres" // cloudsqlpostgres
	_ "github.com/lib/pq"                                                    // postgres
)

// Top-level flags.
var (
	driver = flag.String("driver", "postgres", "database driver")
	conn   = flag.String("conn", "", "database connection string")
)

func main() {
	logger := lg.Default()
	base := command.NewBase(logger)

	// Database commands.
	subcommands.Register(NewMigrate(base), "database admin")

	subcommands.Register(NewIngest(base), "data ingestion")
	subcommands.Register(NewCommits(base), "data ingestion")

	// Help.
	subcommands.Register(subcommands.HelpCommand(), "help")
	subcommands.Register(subcommands.CommandsCommand(), "help")
	subcommands.Register(subcommands.FlagsCommand(), "help")

	// Execute.
	flag.Parse()
	ctx := command.BackgroundContext(logger)
	os.Exit(int(subcommands.Execute(ctx)))
}

// open database connection.
func open() (*sql.DB, error) {
	return sql.Open(*driver, *conn)
}
