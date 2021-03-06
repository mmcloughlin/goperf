package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/app/coordinator"
	"github.com/mmcloughlin/goperf/app/db"
	"github.com/mmcloughlin/goperf/app/sched"
	"github.com/mmcloughlin/goperf/internal/errutil"
	"github.com/mmcloughlin/goperf/pkg/command"
	"github.com/mmcloughlin/goperf/pkg/fs"
)

func main() {
	command.RunError(run)
}

var (
	addr = flag.String("http", "localhost:5050", "http address")
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

	// Build coordinator.
	scheduler := sched.NewDefault(d)
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
