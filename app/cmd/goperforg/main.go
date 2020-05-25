package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/app/dashboard"
	"github.com/mmcloughlin/goperf/app/db"
	"github.com/mmcloughlin/goperf/app/gcs"
	"github.com/mmcloughlin/goperf/internal/errutil"
	"github.com/mmcloughlin/goperf/pkg/command"
	"github.com/mmcloughlin/goperf/pkg/fs"
)

func main() {
	command.RunError(run)
}

var (
	addr    = flag.String("http", "localhost:6060", "http address")
	tmpl    = flag.String("templates", "", "templates directory")
	static  = flag.String("static", "", "static assets directory")
	conn    = flag.String("conn", "", "database connection string")
	bucket  = flag.String("bucket", "", "data files bucket")
	nocache = flag.Bool("nocache", false, "disable asset caches")
)

func run(ctx context.Context, l *zap.Logger) (err error) {
	flag.Parse()

	// Open database connection.
	d, err := db.Open(ctx, *conn)
	if err != nil {
		return err
	}
	defer errutil.CheckClose(&err, d)

	// Build handlers.
	opts := []dashboard.Option{dashboard.WithLogger(l)}

	if *bucket != "" {
		datafs, err := gcs.New(ctx, *bucket)
		if err != nil {
			return err
		}
		opts = append(opts, dashboard.WithDataFileSystem(datafs))
	}

	if *tmpl != "" {
		templates := dashboard.NewTemplates(fs.NewLocal(*tmpl))
		templates.SetCacheEnabled(!*nocache)
		opts = append(opts, dashboard.WithTemplates(templates))
	}

	if *static != "" {
		opts = append(opts, dashboard.WithStaticFileSystem(fs.NewLocal(*static)))
	}

	h := dashboard.NewHandlers(d, opts...)

	if err := h.Init(ctx); err != nil {
		return err
	}

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
