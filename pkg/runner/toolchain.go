package runner

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/build/buildenv"
	"golang.org/x/build/dashboard"

	"github.com/mmcloughlin/cb/pkg/cfg"
	"github.com/mmcloughlin/cb/pkg/lg"
)

type Toolchain interface {
	// Type identifier.
	Type() string
	// String gives a concise identifier for the toolchain.
	String() string
	// Ref is a git reference to the Go repository version (sha or tag).
	Ref() string
	// Configuration returns configuration lines for the
	Configuration() (cfg.Configuration, error)
	// Install the toolchain to the given location in the workspace.
	Install(w *Workspace, root string)
}

func NewToolchain(typ string, params map[string]string) (Toolchain, error) {
	// Define toolchain types.
	constructors := map[string]struct {
		Fields   []string
		Defaults map[string]string
		Make     func(map[string]string) (Toolchain, error)
	}{
		"snapshot": {
			Fields: []string{"builder_type", "revision"},
			Make: func(params map[string]string) (Toolchain, error) {
				return NewSnapshot(params["builder_type"], params["revision"]), nil
			},
		},
		"release": {
			Fields: []string{"version", "os", "arch"},
			Defaults: map[string]string{
				"os":   runtime.GOOS,
				"arch": runtime.GOARCH,
			},
			Make: func(params map[string]string) (Toolchain, error) {
				return NewRelease(params["version"], params["os"], params["arch"]), nil
			},
		},
	}

	// Lookup constructor.
	c, ok := constructors[typ]
	if !ok {
		return nil, fmt.Errorf("unknown toolchain type: %q", typ)
	}

	// Apply defaults.
	params = merge(c.Defaults, params)

	// Ensure required fields are defined.
	var missing []string
	for _, field := range c.Fields {
		if _, ok := params[field]; !ok {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing parameter%s: %s", plural(missing), strings.Join(missing, ", "))
	}

	// Check for extra fields.
	if len(params) > len(c.Fields) {
		fieldset := map[string]bool{}
		for _, field := range c.Fields {
			fieldset[field] = true
		}

		var extra []string
		for field := range params {
			if !fieldset[field] {
				extra = append(extra, field)
			}
		}

		sort.Strings(extra)
		return nil, fmt.Errorf("unknown parameter%s: %s", plural(extra), strings.Join(extra, ", "))
	}

	// Construct.
	return c.Make(params)
}

// merge maps from left to right.
func merge(ms ...map[string]string) map[string]string {
	merged := map[string]string{}
	for _, m := range ms {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

func plural(collection []string) string {
	if len(collection) > 1 {
		return "s"
	}
	return ""
}

// ToolchainConfigurationProvider provides benchmark configuration lines about the given toolchain.
func ToolchainConfigurationProvider(tc Toolchain) cfg.Provider {
	return cfg.NewProviderFunc("toolchain", "go toolchain properties", func() (cfg.Configuration, error) {
		c, err := tc.Configuration()
		if err != nil {
			return nil, err
		}
		c = append(c,
			cfg.Property("type", "toolchain type", cfg.StringValue(tc.Type())),
			cfg.Property("ref", "git reference", cfg.StringValue(tc.Ref())),
		)
		return c, nil
	})
}

type snapshot struct {
	buildertype string
	rev         string
}

func NewSnapshot(buildertype, rev string) Toolchain {
	return &snapshot{
		buildertype: buildertype,
		rev:         rev,
	}
}

func (s *snapshot) Type() string { return "snapshot" }

func (s *snapshot) String() string {
	return path.Join(s.Type(), s.buildertype, s.rev)
}

func (s *snapshot) Ref() string {
	return s.rev
}

func (s *snapshot) Configuration() (cfg.Configuration, error) {
	return cfg.Configuration{
		cfg.Property("buildertype", "snapshot builder type", cfg.StringValue(s.buildertype)),
		cfg.Property("revision", "snapshot revision", cfg.StringValue(s.rev)),
	}, nil
}

func (s *snapshot) Install(w *Workspace, root string) {
	defer lg.Scope(w.Log, "snapshot_install")()

	// Determine download URL.
	url := buildenv.Production.SnapshotURL(s.buildertype, s.rev)

	w.Log.Info("install snapshot",
		zap.String("builder_type", s.buildertype),
		zap.String("go_revision", s.rev),
		zap.String("snapshot_url", url),
	)

	// Download.
	dldir := w.Sandbox("dl")
	archive := filepath.Join(dldir, "go.tar.gz")
	w.Download(url, archive)

	// Extract.
	w.Uncompress(archive, root)
}

// SnapshotBuilderType looks for a suitable builder type to download snapshots
// for the given GOOS/GOARCH.
func SnapshotBuilderType(goos, goarch string) (string, bool) {
	builder, found := "", false
	for name, conf := range dashboard.Builders {
		if conf.SkipSnapshot || conf.GOOS() != goos || conf.GOARCH() != goarch || conf.IsRace() {
			continue
		}
		if !found || len(name) < len(builder) {
			builder = name
			found = true
		}
	}
	return builder, found
}

type release struct {
	os      string
	arch    string
	version string
}

// NewRelease constructs a release toolchain for the given version, os and
// architecture. Version is expected to begin with "go", for example "go1.13.4".
func NewRelease(version, os, arch string) Toolchain {
	return &release{
		version: version,
		os:      os,
		arch:    arch,
	}
}

func (r *release) Type() string { return "release" }

func (r *release) String() string {
	return path.Join(r.Type(), r.version, r.os, r.arch)
}

func (r *release) Ref() string {
	return r.version
}

func (r *release) Configuration() (cfg.Configuration, error) {
	return cfg.Configuration{
		cfg.Property("os", "release operating system", cfg.StringValue(r.os)),
		cfg.Property("arch", "release architecture", cfg.StringValue(r.arch)),
		cfg.Property("version", "release version", cfg.StringValue(r.version)),
	}, nil
}

func (r *release) Install(w *Workspace, root string) {
	defer lg.Scope(w.Log, "release_install")()

	// Determine download URL.
	// TODO(mbm): fetch files list in json format
	const base = "https://golang.org/dl/"
	filename := fmt.Sprintf("%s.%s-%s.tar.gz", r.version, r.os, r.arch)
	url := base + filename

	w.Log.Info("install release",
		zap.String("version", r.version),
		zap.String("os", r.os),
		zap.String("arch", r.arch),
		zap.String("url", url),
	)

	// Download.
	dldir := w.Sandbox("dl")
	archive := filepath.Join(dldir, filename)
	w.Download(url, archive)

	// Extract.
	w.Uncompress(archive, dldir)
	extracted := filepath.Join(dldir, "go")

	// Move into place.
	w.Move(extracted, root)
}
