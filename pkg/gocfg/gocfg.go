// Package gocfg provides configuration related to the Go runtime.
package gocfg

import (
	"os"
	"runtime"

	"github.com/mmcloughlin/goperf/pkg/cfg"
)

// Provider for Go runtime information.
var Provider = cfg.NewProviderFunc("go", "go runtime configuration", gocfg)

func gocfg() (cfg.Configuration, error) {
	return cfg.Configuration{
		cfg.PerfProperty("os", "benchmark runner go operating system target", cfg.StringValue(runtime.GOOS)),
		cfg.PerfProperty("arch", "benchmark runner go architecture target", cfg.StringValue(runtime.GOARCH)),
		cfg.PerfProperty("gc", "GOGC environment variable", cfg.StringValue(os.Getenv("GOGC"))),
		cfg.PerfProperty("debug", "GODEBUG environment variable", cfg.StringValue(os.Getenv("GODEBUG"))),
	}, nil
}
