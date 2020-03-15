// Package meta provides versioning information.
package meta

import (
	"time"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

const placeholder = "unknown"

// Static project information.
var (
	Name = "cb"
)

// Version and build information. Populated at build time.
var (
	Version   = placeholder
	GitSHA    = placeholder
	BuildTime = placeholder
)

// Populated returns whether build information has been populated.
func Populated() bool {
	return Version != placeholder
}

// Provider provides configuration based on the project version.
type Provider struct{}

// Key returns "meta".
func (Provider) Key() cfg.Key { return "meta" }

// Doc describes this configuration provider.
func (Provider) Doc() string { return "project version and build information" }

// Available checks whether version information was populated.
func (Provider) Available() bool { return Populated() }

// Configuration returns project version and build information.
func (Provider) Configuration() (cfg.Configuration, error) {
	t, err := time.Parse("2006-01-02T15:04:05-0700", BuildTime)
	if err != nil {
		return nil, err
	}
	return cfg.Configuration{
		cfg.Property("name", "project name", cfg.StringValue(Name)),
		cfg.Property("version", "project version", cfg.StringValue(Version)),
		cfg.Property("gitsha", "git sha", cfg.StringValue(GitSHA)),
		cfg.Property("buildtime", "build time", cfg.TimeValue(t)),
	}, nil
}
