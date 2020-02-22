package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/mmcloughlin/cb/internal/flags"
	"github.com/mmcloughlin/cb/pkg/cfg"
	"github.com/mmcloughlin/cb/pkg/sys"
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
	var keys flags.Strings = providers.Keys()
	flag.Var(&keys, "cfg", "config types to include")
	flag.Parse()

	args := flag.Args()

	// Determine the config providers.
	ps, err := providers.Select(keys...)
	if err != nil {
		return err
	}

	c, err := ps.Configuration()
	if err != nil {
		return err
	}

	if err := cfg.Write(os.Stdout, c); err != nil {
		return err
	}

	// Execute the sub-process.
	return execute(args)
}

// providers is a list of all supported config sources.
var providers = cfg.Providers{
	sys.Host,
	sys.LoadAverage,
	sys.VirtualMemory,
	sys.CPU,
	sys.Caches{},
	sys.ProcStat{},
	sys.CPUFreq{},
	sys.IntelPState{},
}

func execute(args []string) error {
	if len(args) == 0 {
		return errors.New("no command provided")
	}
	argv0, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	return syscall.Exec(argv0, args, os.Environ())
}
