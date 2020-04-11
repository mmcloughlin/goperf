package worker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/mmcloughlin/cb/app/coordinator"
	"github.com/mmcloughlin/cb/internal/errutil"
	"github.com/mmcloughlin/cb/pkg/lg"
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
	log       lg.Logger

	queue []*coordinator.Job
}

type Option func(*Worker)

func New(c *coordinator.Client, p Processor, opts ...Option) *Worker {
	return &Worker{
		client:    c,
		processor: p,
		poll:      DefaultPollingConfig,
		log:       lg.Noop(),
	}
}

func WithPollingConfig(poll PollingConfig) Option {
	return func(w *Worker) { w.poll = poll }
}

func WithLogger(l lg.Logger) Option {
	return func(w *Worker) { w.log = l }
}

func (w *Worker) Run(ctx context.Context) error {
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
			lg.Error(w.log, "job processing", err)
		}
	}
}

// next polls indefinitely for more work.
func (w *Worker) next(ctx context.Context) (*coordinator.Job, error) {
	interval := w.poll.Initial

	for len(w.queue) == 0 {
		res, err := w.client.Jobs(ctx)
		if err == nil && len(res.Jobs) > 0 {
			w.queue = append(w.queue, res.Jobs...)
			break
		}
		if err != nil {
			lg.Error(w.log, "jobs request", err)
		}

		// Sleep before polling again.
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
	if err := w.client.Fail(ctx, j.UUID); err != nil {
		lg.Error(w.log, "reporting job failure", err)
	}
}

func (w *Worker) halt(ctx context.Context, j *coordinator.Job) {
	if err := w.client.Halt(ctx, j.UUID); err != nil {
		lg.Error(w.log, "halting job", err)
	}
}
