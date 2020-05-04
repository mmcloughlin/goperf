package coordinator

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/app/dashboard"
	"github.com/mmcloughlin/cb/app/httputil"
	"github.com/mmcloughlin/cb/app/service"
)

// Initialization.
var handler *dashboard.Handlers

func initialize(ctx context.Context, l *zap.Logger) error {
	// Database.
	d, err := service.DB(ctx, l)
	if err != nil {
		return err
	}

	// Data filesystem.
	datafs, err := service.ResultsFileSystem(ctx)
	if err != nil {
		return err
	}

	// Handlers.
	handler = dashboard.NewHandlers(d,
		dashboard.WithLogger(l),
		dashboard.WithDataFileSystem(datafs),
		dashboard.WithCacheControl(httputil.CacheControl{
			MaxAge:     10 * time.Minute,
			Directives: []string{"public"},
		}),
	)

	if err := handler.Init(ctx); err != nil {
		return err
	}

	return nil
}

func init() {
	service.Initialize(initialize)
}

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}
