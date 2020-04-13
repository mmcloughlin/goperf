// Package command provides a common structure for building command line programs.
package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"

	"github.com/google/subcommands"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mmcloughlin/cb/pkg/sig"
)

// MainStatus is an entry point returnint an exit status.
type MainStatus func(context.Context, *zap.Logger) int

func Run(m MainStatus) {
	// Initialize logger.
	l, err := Logger()
	if err != nil {
		log.Fatalf("logger initialization: %s", err)
	}

	// Context.
	ctx := BackgroundContext(l)

	os.Exit(m(ctx, l))
}

// MainError is an entry point returning an error.
type MainError func(context.Context, *zap.Logger) error

func RunError(m MainError) {
	Run(func(ctx context.Context, l *zap.Logger) int {
		if err := m(ctx, l); err != nil {
			l.Error("execution error", zap.Error(err))
			return 1
		}
		return 0
	})
}

// Logger initializes a logger suitable for command-line applications.
func Logger() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	cfg.EncoderConfig.CallerKey = ""
	return cfg.Build()
}

// BackgroundContext returns a context suitable for a command-line tool or service.
func BackgroundContext(log *zap.Logger) context.Context {
	return sig.ContextWithSignal(context.Background(), func(s os.Signal) {
		log.Info("cancelling on signal", zap.Stringer("signal", s))
	}, syscall.SIGINT, syscall.SIGTERM)
}

// Base is a base for all subcommands.
type Base struct {
	Log *zap.Logger
}

// NewBase builds a new base command for the named tool.
func NewBase(l *zap.Logger) Base {
	return Base{
		Log: l,
	}
}

// SetFlags is a stub implementation of the SetFlags methods that does nothing.
func (Base) SetFlags(f *flag.FlagSet) {}

// UsageError logs a usage error and returns a suitable exit code.
func (b Base) UsageError(format string, args ...interface{}) subcommands.ExitStatus {
	b.Log.Info(fmt.Sprintf(format, args...))
	return subcommands.ExitUsageError
}

// Fail logs an error message and returns a failing exit code.
func (b Base) Fail(format string, args ...interface{}) subcommands.ExitStatus {
	b.Log.Error(fmt.Sprintf(format, args...))
	return subcommands.ExitFailure
}

// Error logs err and returns a failing exit code.
func (b Base) Error(err error) subcommands.ExitStatus {
	return b.Fail(err.Error())
}

// Status fails if the provided error is non-nil, and returns success otherwise.
func (b Base) Status(err error) subcommands.ExitStatus {
	if err != nil {
		return b.Error(err)
	}
	return subcommands.ExitSuccess
}

// CheckClose closes c. On error it logs and writes to the status pointer.
// Intended for deferred Close() calls.
func (b Base) CheckClose(statusp *subcommands.ExitStatus, c io.Closer) {
	if err := c.Close(); err != nil {
		*statusp = b.Error(err)
	}
}
