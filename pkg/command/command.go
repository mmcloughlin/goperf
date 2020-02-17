package command

import (
	"context"
	"os"
	"syscall"

	"github.com/mmcloughlin/cb/pkg/lg"
	"github.com/mmcloughlin/cb/pkg/sig"
)

// BackgroundContext returns a context suitable for a command-line tool or service.
func BackgroundContext(l lg.Logger) context.Context {
	return sig.ContextWithSignal(context.Background(), func(s os.Signal) {
		l.Printf("received %s: cancelling")
	}, syscall.SIGINT, syscall.SIGTERM)
}
