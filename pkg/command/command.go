package command

import (
	"context"
	"flag"
	"os"
	"syscall"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/pkg/lg"
	"github.com/mmcloughlin/cb/pkg/sig"
)

// BackgroundContext returns a context suitable for a command-line tool or service.
func BackgroundContext(l lg.Logger) context.Context {
	return sig.ContextWithSignal(context.Background(), func(s os.Signal) {
		l.Printf("received %s: cancelling", s)
	}, syscall.SIGINT, syscall.SIGTERM)
}

// Base is a base for all subcommands.
type Base struct {
	Log lg.Logger
}

// NewBase builds a new base command for the named tool.
func NewBase(l lg.Logger) Base {
	return Base{
		Log: l,
	}
}

// SetFlags is a stub implementation of the SetFlags methods that does nothing.
func (Base) SetFlags(f *flag.FlagSet) {}

// UsageError logs a usage error and returns a suitable exit code.
func (b Base) UsageError(format string, args ...interface{}) subcommands.ExitStatus {
	b.Log.Printf(format, args...)
	return subcommands.ExitUsageError
}

// Fail logs an error message and returns a failing exit code.
func (b Base) Fail(format string, args ...interface{}) subcommands.ExitStatus {
	b.Log.Printf(format, args...)
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
