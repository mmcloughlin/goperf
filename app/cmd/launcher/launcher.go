package main

import (
	"log"
	"os"
	"time"

	"github.com/mmcloughlin/cb/app/launch"
	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/job"
	"github.com/mmcloughlin/cb/pkg/lg"
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
	logger := lg.Default()
	ctx := command.BackgroundContext(logger)

	// Create launcher.
	l, err := launch.NewLauncher(ctx, topic)
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
				"builder_type": "linux-amd64",
				"revision":     "60d437f99468906935f35e5c6fbd31c7228a1045",
			},
		},
		Suites: []job.Suite{
			{
				Module: job.Module{
					Path:    "github.com/klauspost/compress",
					Version: "b949da471e55fbe4393e6eb595602d936f5c312e",
				},
				Short:     true,
				BenchTime: 5 * time.Millisecond,
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
