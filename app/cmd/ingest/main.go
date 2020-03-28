// Command ingest populates a database with benchmark results.
package main

import (
	"flag"
	"os"

	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/app/gcs"
	"github.com/mmcloughlin/cb/app/results"
	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/lg"
)

func main() {
	os.Exit(main1())
}

func main1() int {
	l := lg.Default()
	if err := mainerr(l); err != nil {
		l.Printf("error: %s", err)
		return 1
	}
	return 0
}

var (
	bucket = flag.String("bucket", "", "data files bucket")
	conn   = flag.String("conn", "", "database connection string")
)

func mainerr(l lg.Logger) error {
	flag.Parse()

	ctx := command.BackgroundContext(l)

	// Open filesystem.
	fs, err := gcs.New(ctx, *bucket)
	if err != nil {
		return err
	}

	// Open database connection.
	d, err := db.Open(*conn)
	if err != nil {
		return err
	}
	defer d.Close()

	// List files.
	files, err := fs.List(ctx, "")
	if err != nil {
		return err
	}

	// Extract results.
	loader, err := results.NewLoader(results.WithFilesystem(fs))
	if err != nil {
		return err
	}

	for _, file := range files {
		l.Printf("file=%s mod=%s", file.Path, file.ModTime)

		rs, err := loader.Load(ctx, file.Path)
		if err != nil {
			l.Printf("loading error: %s", err)
			continue
		}
		l.Printf("loaded %d results", len(rs))

		for _, r := range rs {
			if err := d.StoreResult(ctx, r); err != nil {
				return err
			}
		}
		l.Printf("inserted %d results", len(rs))
	}

	return nil
}
