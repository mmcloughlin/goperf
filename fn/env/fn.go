package env

import (
	"context"
	"net/http"
	"os"

	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/app/httputil"
	"github.com/mmcloughlin/goperf/app/service"
)

// Initialization.
var logger *zap.Logger

func initialize(ctx context.Context, l *zap.Logger) error {
	logger = l
	return nil
}

func init() {
	service.Initialize(initialize)
}

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	// Log environment.
	for _, e := range os.Environ() {
		logger.Info(e)
	}

	// Report ok.
	httputil.OK(w)
}
