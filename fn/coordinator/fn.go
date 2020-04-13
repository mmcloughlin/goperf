package coordinator

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/app/coordinator"
	"github.com/mmcloughlin/cb/app/sched"
	"github.com/mmcloughlin/cb/app/service"
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
	pri := sched.TimeSinceSmoothStep(
		60*24*time.Hour, sched.PriorityHigh,
		365*24*time.Hour, sched.PriorityIdle,
	)
	scheduler := sched.NewRecentCommits(d, pri)
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
