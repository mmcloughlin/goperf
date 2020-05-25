package coordinator

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/app/coordinator"
	"github.com/mmcloughlin/goperf/app/sched"
	"github.com/mmcloughlin/goperf/app/service"
)

// Initialization.
var handler *coordinator.Handlers

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

	// Coordinator.
	scheduler := sched.NewDefault(d)
	coord := coordinator.New(d, scheduler, datafs)

	// Handlers.
	handler = coordinator.NewHandlers(coord, l)

	return nil
}

func init() {
	service.Initialize(initialize)
}

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}
