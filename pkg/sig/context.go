// Package sig provides signal helpers.
package sig

import (
	"context"
	"os"
	"os/signal"
)

// ContextWithSignal returns a context that will be cancelled on any of the
// provided signals. Calls pre function before cancellation, if non-nil.
func ContextWithSignal(parent context.Context, pre func(os.Signal), sigs ...os.Signal) context.Context {
	ctx, cancel := context.WithCancel(parent)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)

	go func() {
		select {
		case <-ctx.Done():
		case s := <-ch:
			if pre != nil {
				pre(s)
			}
			cancel()
		}
		signal.Stop(ch)
	}()

	return ctx
}
