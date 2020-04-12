package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/app/coordinator"
	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/app/sched"
	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/fs"
)

func main() {
	command.RunError(run)
}

var (
	addr = flag.String("http", "localhost:5050", "http address")
	conn = flag.String("conn", "", "database connection string")
	data = flag.String("data", "", "data directory")
)

func run(ctx context.Context, l *zap.Logger) error {
	flag.Parse()

	// Open database connection.
	d, err := db.Open(ctx, *conn)
	if err != nil {
		return err
	}
	defer d.Close()

	// Build coordinator.
	pri := sched.TimeSinceSmoothStep(
		60*24*time.Hour, sched.PriorityHigh,
		365*24*time.Hour, sched.PriorityIdle,
	)
	scheduler := sched.NewRecentCommits(d, pri)
	datafs := fs.NewLocal(*data)
	c := coordinator.New(d, scheduler, datafs)
	c.SetLogger(l)

	// Build coordinator handlers.
	h := coordinator.NewHandlers(c, l)

	// Launch server.
	s := &http.Server{
		Addr:        *addr,
		Handler:     h,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	errc := make(chan error)
	go func() {
		errc <- s.ListenAndServe()
	}()

	// Wait for context cancellation or error from server.
	select {
	case <-ctx.Done():
	case err := <-errc:
		return err
	}

	// Shutdown server.
	l.Info("http server shutdown")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.Shutdown(ctx)
}
