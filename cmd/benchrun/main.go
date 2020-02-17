package main

import (
	"flag"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/mmcloughlin/cb/internal/flags"
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
	)
	flag.Var(&toolchainconfig, "toolchain", "toolchain configuration")
	flag.Parse()

	// Build toolchain.
	tc, err := runner.NewToolchain(toolchainconfig.Type, toolchainconfig.Params.Map())
	if err != nil {
		return err
	}

	// Construct workspace.
	// TODO(mbm): get it working without inheriting environment
	w, err := runner.NewWorkspace(runner.InheritEnviron())
	if err != nil {
		return err
	}

	// Initialize runner.
	r := runner.NewRunner(w, tc)
	if err := r.Init(); err != nil {
		return err
	}

	// Run benchmark.
	mod := runner.Module{
		Path:    path.Join("github.com", owner, repo),
		Version: rev,
	}
	job := runner.Job{
		Module: mod,
	}
	if err := r.Benchmark(job); err != nil {
		return err
	}

	return nil
}
