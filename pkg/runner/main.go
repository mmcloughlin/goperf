package main

import (
	"log"
	"os"
	"path"
	"runtime"
)

const (
	gorevision  = "60d437f99468906935f35e5c6fbd31c7228a1045"
	buildertype = "linux-amd64" // "darwin-amd64-10_14"

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
	// TODO(mbm): get it working without inheriting environment
	w, err := NewWorkspace(InheritEnviron())
	if err != nil {
		return err
	}

	// tc := NewSnapshot(buildertype, gorevision)
	tc := NewRelease("1.13.8", runtime.GOOS, runtime.GOARCH)
	r := NewRunner(w, tc)

	if err := r.Init(); err != nil {
		return err
	}

	mod := Module{
		Path:    path.Join("github.com", owner, repo),
		Version: rev,
	}
	job := Job{
		Module: mod,
	}
	r.Benchmark(job)

	return nil
}
