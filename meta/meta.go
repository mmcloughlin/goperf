// Package meta provides versioning information.
package meta

import (
	"github.com/mmcloughlin/goperf/pkg/cfg"
)

const placeholder = "unknown"

// Static project information.
var (
	Name = "goperf"
)

// Version and build information. Populated at build time.
var (
	Version = placeholder
	GitSHA  = placeholder
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
	return cfg.Configuration{
		cfg.Property("name", "project name", cfg.StringValue(Name)),
		cfg.Property("version", "project version", cfg.StringValue(Version)),
		cfg.Property("gitsha", "git sha", cfg.StringValue(GitSHA)),
	}, nil
}
