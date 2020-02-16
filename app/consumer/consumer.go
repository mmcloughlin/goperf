// Package consumer implements a pubsub consumer.
package consumer

import (
	"context"
	"time"

	pubsub "cloud.google.com/go/pubsub/apiv1"
	pubsubpb "google.golang.org/genproto/googleapis/pubsub/v1"

	"github.com/mmcloughlin/cb/pkg/lg"
)

type Handler interface {
	Handle(context.Context, []byte) error
}

type HandlerFunc func(context.Context, []byte) error

func (f HandlerFunc) Handle(ctx context.Context, data []byte) error {
	return f(ctx, data)
}

type Consumer struct {
	client  *pubsub.SubscriberClient
	sub     string
	handler Handler

	extend time.Duration
	grace  time.Duration

	l lg.Logger
}

var defaultconsumer = Consumer{
	extend: 30 * time.Second,
	grace:  5 * time.Second,
}

type Option func(*Consumer)

func New(ctx context.Context, sub string, h Handler, opts ...Option) (*Consumer, error) {
	// Build client.
	client, err := pubsub.NewSubscriberClient(ctx)
	if err != nil {
		return nil, err
	}

	// Populate consumer.
	c := &Consumer{}
	*c = defaultconsumer
	c.client = client
	c.sub = sub
	c.handler = h
	c.l = lg.Default()

	// Custom options.
	c.Options(opts...)

	return c, nil
}

func (c *Consumer) Options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

func WithExtensionPeriod(d time.Duration) Option {
	return func(c *Consumer) { c.extend = d }
}

func WithLogger(l lg.Logger) Option {
	return func(c *Consumer) { c.l = l }
}

func (c *Consumer) Close() error {
	return c.client.Close()
}

func (c *Consumer) Receive(ctx context.Context) error {
	defer lg.Scope(c.l, "consumer_receive_loop")()
	for {
		if err := c.receive(ctx); err != nil {
			return err
		}
	}
}

func (c *Consumer) receive(ctx context.Context) error {
	defer lg.Scope(c.l, "consumer_receive")()

	// Pull message.
	m, err := c.pull(ctx)
	if err != nil {
		return err
	}

	if m == nil {
		return nil
	}

	c.l.Printf("received message")
	lg.Param(c.l, "message id", m.Message.MessageId)

	// Process the message in a goroutine.
	hctx, hcancel := context.WithCancel(ctx)
	errc := make(chan error)
	go func() {
		errc <- c.handler.Handle(hctx, m.Message.Data)
	}()

	// Extend the deadline while we're waiting for the result.
	delay := 0 * time.Second
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errc:
			if err != nil {
				// TODO(mbm): nack
				return err
			}
			return c.ack(ctx, m)
		case <-time.After(delay):
			if err := c.extenddeadline(ctx, m, c.extend); err != nil {
				hcancel()
				return err
			}
			delay = c.extend - c.grace
		}
	}
}

// pull message from subscription.
func (c *Consumer) pull(ctx context.Context) (*pubsubpb.ReceivedMessage, error) {
	req := &pubsubpb.PullRequest{
		Subscription: c.sub,
		MaxMessages:  1,
	}

	res, err := c.client.Pull(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(res.ReceivedMessages) == 0 {
		return nil, nil
	}

	return res.ReceivedMessages[0], nil
}

// ack message.
func (c *Consumer) ack(ctx context.Context, m *pubsubpb.ReceivedMessage) error {
	return c.client.Acknowledge(ctx, &pubsubpb.AcknowledgeRequest{
		Subscription: c.sub,
		AckIds:       []string{m.AckId},
	})
}

// extenddeadline extends the ack deadline for m by extend duration.
func (c *Consumer) extenddeadline(ctx context.Context, m *pubsubpb.ReceivedMessage, extend time.Duration) error {
	return c.client.ModifyAckDeadline(ctx, &pubsubpb.ModifyAckDeadlineRequest{
		Subscription:       c.sub,
		AckIds:             []string{m.AckId},
		AckDeadlineSeconds: int32(extend.Seconds()),
	})
}
