package job

import (
	"context"
	"fmt"

	pubsub "cloud.google.com/go/pubsub/apiv1"
	pubsubpb "google.golang.org/genproto/googleapis/pubsub/v1"
)

type Submission struct {
	Job *Job
	ID  string
}

type Launcher struct {
	client *pubsub.PublisherClient
	topic  string
}

func NewLauncher(ctx context.Context, topic string) (*Launcher, error) {
	client, err := pubsub.NewPublisherClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Launcher{
		client: client,
		topic:  topic,
	}, nil
}

func (l *Launcher) Close() error {
	return l.client.Close()
}

func (l *Launcher) Launch(ctx context.Context, j *Job) (*Submission, error) {
	submissions, err := l.Batch(ctx, []*Job{j})
	if err != nil {
		return nil, err
	}
	if len(submissions) != 1 {
		panic("expect one submission")
	}
	return submissions[0], nil
}

func (l *Launcher) Batch(ctx context.Context, jobs []*Job) ([]*Submission, error) {
	if len(jobs) == 0 {
		return nil, nil
	}

	// Build messages.
	msgs := make([]*pubsubpb.PubsubMessage, 0, len(jobs))
	for _, j := range jobs {
		data, err := Marshal(j)
		if err != nil {
			return nil, fmt.Errorf("marshal job: %w", err)
		}
		msgs = append(msgs, &pubsubpb.PubsubMessage{
			Data: data,
		})
	}

	// Publish.
	req := &pubsubpb.PublishRequest{
		Topic:    l.topic,
		Messages: msgs,
	}

	res, err := l.client.Publish(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("jobs pubsub publish: %w", err)
	}

	// Extract message IDs.
	submissions := make([]*Submission, len(res.MessageIds))
	for i, id := range res.MessageIds {
		submissions[i] = &Submission{
			Job: jobs[i],
			ID:  id,
		}
	}

	return submissions, nil
}
