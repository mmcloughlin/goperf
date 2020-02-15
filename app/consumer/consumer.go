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

func (c *Consumer) receive(ctx context.Context) (err error) {
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

	// Start notification goroutine.
	ctx, cancel := context.WithCancel(ctx)
	errc := make(chan error, 1)
	go c.notify(ctx, errc, m)
	defer func() {
		cancel()
		if err == nil {
			err = <-errc
		}
	}()

	// Process.
	if err := c.handler.Handle(ctx, m.Message.Data); err != nil {
		return err
	}

	// Ack.
	if err := c.ack(ctx, m); err != nil {
		return err
	}

	return nil
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

func (c *Consumer) notify(ctx context.Context, errc chan error, m *pubsubpb.ReceivedMessage) {
	delay := 0 * time.Second
	for {
		select {
		case <-ctx.Done():
			errc <- nil
			return
		case <-time.After(delay):
			err := c.client.ModifyAckDeadline(ctx, &pubsubpb.ModifyAckDeadlineRequest{
				Subscription:       c.sub,
				AckIds:             []string{m.AckId},
				AckDeadlineSeconds: int32(c.extend.Seconds()),
			})
			if err != nil {
				errc <- err
				return
			}
			delay = c.extend - c.grace
		}
	}
}
