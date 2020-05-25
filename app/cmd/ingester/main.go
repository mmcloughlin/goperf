package main

import (
	"context"
	"flag"
	"time"

	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/app/db"
	"github.com/mmcloughlin/goperf/app/entity"
	"github.com/mmcloughlin/goperf/app/ingest"
	"github.com/mmcloughlin/goperf/app/results"
	"github.com/mmcloughlin/goperf/internal/errutil"
	"github.com/mmcloughlin/goperf/pkg/command"
	"github.com/mmcloughlin/goperf/pkg/fs"
)

func main() {
	command.RunError(run)
}

var (
	conn = flag.String("conn", "", "database connection string")
	data = flag.String("data", "", "data directory")
)

func run(ctx context.Context, l *zap.Logger) (err error) {
	flag.Parse()

	// Open database connection.
	d, err := db.Open(ctx, *conn)
	if err != nil {
		return err
	}
	defer errutil.CheckClose(&err, d)

	// Build ingester.
	datafs := fs.NewLocal(*data)
	loader, err := results.NewLoader(results.WithFilesystem(datafs))
	if err != nil {
		return err
	}
	i := ingest.New(d, loader)
	i.SetLogger(l)

	// Enter polling loop.
	interval := time.Second
	for {
		// Look for tasks waiting for ingestion.
		tasks, err := d.ListTasksWithStatus(ctx, []entity.TaskStatus{entity.TaskStatusResultUploaded})
		if err != nil {
			return err
		}

		for _, task := range tasks {
			l.Info("ingest task", zap.Stringer("task_uuid", task.UUID))
			if err := i.Task(ctx, task.UUID); err != nil {
				return err
			}
		}

		// Adjust interval based on whether we got any results.
		if len(tasks) > 0 {
			interval /= 2
			if interval < time.Nanosecond {
				interval = time.Nanosecond
			}
		} else {
			interval *= 2
			if interval > time.Minute {
				interval = time.Minute
			}
		}

		// Sleep.
		l.Debug("wait", zap.Duration("interval", interval))
		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
