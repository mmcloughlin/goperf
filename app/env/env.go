// Package env provides helpers for environment key-value properties.
package env

import (
	"strings"

	"github.com/mmcloughlin/cb/app/entity"
)

// Short identifier of the environment.
func Short(e entity.Properties) string {
	return e["go-arch"]
}

// Title generates a title for the given environment.
func Title(e entity.Properties) string {
	keys := []string{
		"go-os",
		"go-arch",
		"affinecpu-cpu0-modelname",
		"affinecpufreq-cpu0-cpuinfomaxfreq",
	}
	fields := []string{}
	for _, key := range keys {
		if v, ok := e[key]; ok {
			fields = append(fields, v)
		}
	}
	if len(fields) > 0 {
		return strings.Join(fields, ", ")
	}
	return e.UUID().String()
}
