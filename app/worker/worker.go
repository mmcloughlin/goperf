package worker

import (
	"context"
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/app/coordinator"
	"github.com/mmcloughlin/cb/internal/errutil"
)

// Processor executes jobs, returning the output file.
type Processor interface {
	Process(context.Context, *coordinator.Job) (io.ReadCloser, error)
}

type PollingConfig struct {
	Initial    time.Duration
	Multiplier float64
	Max        time.Duration
}

func (p PollingConfig) Next(d time.Duration) time.Duration {
	d = time.Duration(float64(d) * p.Multiplier)
	if d > p.Max {
		d = p.Max
	}
	return d
}

var DefaultPollingConfig = PollingConfig{
	Initial:    time.Second,
	Multiplier: 1.5,
	Max:        time.Minute,
}

type Worker struct {
	client    *coordinator.Client
	processor Processor
	poll      PollingConfig
	log       *zap.Logger

	queue []*coordinator.Job
}

type Option func(*Worker)

func New(c *coordinator.Client, p Processor, opts ...Option) *Worker {
	w := &Worker{
		client:    c,
		processor: p,
		poll:      DefaultPollingConfig,
		log:       zap.NewNop(),
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

func WithPollingConfig(poll PollingConfig) Option {
	return func(w *Worker) { w.poll = poll }
}

func WithLogger(l *zap.Logger) Option {
	return func(w *Worker) { w.log = l.Named("worker") }
}

func (w *Worker) Run(ctx context.Context) error {
	w.log.Info("starting worker loop")

	for {
		// Fetch next job. THe next function polls indefinitely for work, so an
		// error here means the context was cancelled or something else
		// unrecoverable happened. Bail out.
		j, err := w.next(ctx)
		if err != nil {
			return err
		}

		// Process the job. Errors are simply logged and we move onto the next
		// one. The process function will report status to coordinator.
		if err := w.process(ctx, j); err != nil {
			w.log.Error("job processing error", zap.Error(err))
		}
	}
}

// next polls indefinitely for more work.
func (w *Worker) next(ctx context.Context) (*coordinator.Job, error) {
	interval := w.poll.Initial

	for len(w.queue) == 0 {
		w.log.Debug("fetch jobs")

		res, err := w.client.Jobs(ctx)
		if err == nil && len(res.Jobs) > 0 {
			w.queue = append(w.queue, res.Jobs...)
			break
		}
		if err != nil {
			w.log.Error("jobs request error", zap.Error(err))
		}

		// Sleep before polling again.
		w.log.Debug("wait", zap.Duration("interval", interval))

		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		interval = w.poll.Next(interval)
	}

	j := w.queue[0]
	w.queue = w.queue[1:]
	return j, nil
}

func (w *Worker) process(ctx context.Context, j *coordinator.Job) (err error) {
	// Record start of work.
	if err := w.client.Start(ctx, j.UUID); err != nil {
		w.halt(ctx, j)
		return fmt.Errorf("report job start: %w", err)
	}

	// Delegate to worker processor.
	r, err := w.processor.Process(ctx, j)
	if err != nil {
		w.fail(ctx, j)
		return fmt.Errorf("process job: %w", err)
	}
	defer errutil.CheckClose(&err, r)

	// Upload.
	if err := w.client.UploadResult(ctx, j.UUID, r); err != nil {
		w.halt(ctx, j)
		return fmt.Errorf("upload result: %w", err)
	}

	return nil
}

func (w *Worker) fail(ctx context.Context, j *coordinator.Job) {
	w.log.Info("reporting job failure", zap.Stringer("uuid", j.UUID))
	if err := w.client.Fail(ctx, j.UUID); err != nil {
		w.log.Error("error reporting job failure", zap.Error(err))
	}
}

func (w *Worker) halt(ctx context.Context, j *coordinator.Job) {
	w.log.Info("halt job", zap.Stringer("uuid", j.UUID))
	if err := w.client.Halt(ctx, j.UUID); err != nil {
		w.log.Error("error halting job", zap.Error(err))
	}
}
