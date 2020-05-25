package coordinator

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/app/db"
	"github.com/mmcloughlin/goperf/app/httputil"
	"github.com/mmcloughlin/goperf/app/service"
)

// timeout tasks after this duration of inactivity while in a pending state.
const timeout = 6 * time.Hour

// Initialization.
var (
	database *db.DB
	handler  http.Handler
)

func initialize(ctx context.Context, l *zap.Logger) error {
	var err error

	database, err = service.DB(ctx, l)
	if err != nil {
		return err
	}

	handler = httputil.ErrorHandler{
		Handler: httputil.HandlerFunc(handle),
		Log:     l,
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

func handle(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Timeout stale tasks.
	until := time.Now().Add(-timeout)
	if err := database.TimeoutStaleTasks(ctx, until); err != nil {
		return err
	}

	// Report ok.
	httputil.OK(w)

	return nil
}
