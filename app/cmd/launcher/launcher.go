package main

import (
	"context"
	"log"
	"os"

	"github.com/mmcloughlin/cb/app/job"
)

var (
	// TODO(mbm): remove hardcoded topic
	topic = "projects/contbench/topics/jobs"
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

	// Create launcher.
	l, err := job.NewLauncher(ctx, topic)
	if err != nil {
		return err
	}
	defer l.Close()

	// Publish a job.
	// TODO(mbm): remove hardcoded job
	j := &job.Job{
		Toolchain: job.Toolchain{
			Type: "snapshot",
			Params: map[string]string{
				"revision":    "60d437f99468906935f35e5c6fbd31c7228a1045",
				"buildertype": "linux-amd64",
			},
		},
		Suites: []job.Suite{
			{
				Module: job.Module{
					Path:    "github.com/klauspost/compress",
					Version: "b949da471e55fbe4393e6eb595602d936f5c312e",
				},
			},
		},
	}

	s, err := l.Launch(ctx, j)
	if err != nil {
		return err
	}

	log.Printf("submitted job with id = %s", s.ID)

	return nil
}
