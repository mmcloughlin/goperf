package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"

	"github.com/mmcloughlin/cb/app/gcs"
	"github.com/mmcloughlin/cb/app/service"
	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/fs"
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
	addr    = flag.String("http", "localhost:6060", "http address")
	tmpl    = flag.String("templates", "", "templates directory")
	static  = flag.String("static", "", "static assets directory")
	project = flag.String("project", "", "google cloud project")
	bucket  = flag.String("bucket", "", "data files bucket")
)

func mainerr(l lg.Logger) error {
	flag.Parse()

	ctx := command.BackgroundContext(l)

	// Configure firestore backend.
	fsc, err := firestore.NewClient(ctx, *project)
	if err != nil {
		return err
	}
	defer fsc.Close()

	srv := service.NewFirestore(fsc)

	// Build handlers.
	var opts []Option

	if *bucket != "" {
		datafs, err := gcs.New(ctx, *bucket)
		if err != nil {
			return err
		}
		opts = append(opts, WithDataFileSystem(datafs))
	}

	if *tmpl != "" {
		opts = append(opts, WithTemplateFileSystem(fs.NewLocal(*tmpl)))
	}

	if *static != "" {
		opts = append(opts, WithStaticFileSystem(fs.NewLocal(*static)))
	}

	h := NewHandlers(srv, opts...)

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
	l.Printf("http server shutdown")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.Shutdown(ctx)
}
