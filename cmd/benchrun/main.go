package main

import (
	"flag"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/mmcloughlin/cb/internal/flags"
	"github.com/mmcloughlin/cb/pkg/fs"
	"github.com/mmcloughlin/cb/pkg/runner"
)

const (
	owner = "klauspost"
	repo  = "compress"
	rev   = "b949da471e55fbe4393e6eb595602d936f5c312e"
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
	// Flags.
	var (
		toolchainconfig = flags.TypeParams{
			Type: "release",
			Params: flags.Params{
				{Key: "version", Value: runtime.Version()},
			},
		}
		output   string
		preserve bool
	)

	flag.Var(&toolchainconfig, "toolchain", "toolchain configuration")
	flag.StringVar(&output, "output", "", "output path")
	flag.BoolVar(&preserve, "preserve", false, "preserve working directory")

	flag.Parse()

	// Build toolchain.
	tc, err := runner.NewToolchain(toolchainconfig.Type, toolchainconfig.Params.Map())
	if err != nil {
		return err
	}

	// Construct workspace.
	w, err := runner.NewWorkspace()
	if err != nil {
		return err
	}

	if output != "" {
		w.Options(runner.WithArtifactStore(fs.NewLocal(output)))
	}

	// Initialize runner.
	r := runner.NewRunner(w, tc)
	r.Init()

	// Run benchmark.
	mod := runner.Module{
		Path:    path.Join("github.com", owner, repo),
		Version: rev,
	}
	job := runner.Job{
		Module: mod,
	}
	r.Benchmark(job)

	// Clean.
	if !preserve {
		r.Clean()
	}

	return w.Error()
}
