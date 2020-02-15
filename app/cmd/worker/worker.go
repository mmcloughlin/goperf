package main

import (
	"context"
	"log"
	"os"

	"github.com/mmcloughlin/cb/app/consumer"
)

var (
	subscription = "projects/contbench/subscriptions/worker_jobs"
)

func main() {
	os.Exit(main1())
}

func main1() int {
	if err := mainerr(); err != nil {
		log.Print(err)
		return 1
	}
	return 0
}

func mainerr() error {
	ctx := context.Background()
	c, err := consumer.New(ctx, subscription, consumer.HandlerFunc(func(ctx context.Context, data []byte) error {
		log.Printf("data = %s", data)
		return nil
	}))
	if err != nil {
		return err
	}
	defer c.Close()
	return c.Receive(ctx)
}
