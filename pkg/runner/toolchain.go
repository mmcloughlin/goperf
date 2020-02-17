package runner

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"golang.org/x/build/buildenv"

	"github.com/mmcloughlin/cb/pkg/lg"
)

type Toolchain interface {
	// TODO(mbm): Toolchain returns configuration lines rather than a string
	String() string
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

func (s *snapshot) String() string {
	return path.Join("snapshot", s.buildertype, s.rev)
}

func (s *snapshot) Install(w *Workspace, root string) {
	defer lg.Scope(w, "snapshot_install")()

	// Log parameters.
	lg.Param(w, "snapshot_builder_type", s.buildertype)
	lg.Param(w, "snapshot_go_revision", s.rev)

	// Determine download URL.
	url := buildenv.Production.SnapshotURL(s.buildertype, s.rev)
	lg.Param(w, "snapshot_url", url)

	// Download.
	dldir := w.Sandbox("dl")
	archive := filepath.Join(dldir, "go.tar.gz")
	w.Download(url, archive)

	// Extract.
	w.Uncompress(archive, root)
}

type release struct {
	os      string
	arch    string
	version string
}

func NewRelease(version, os, arch string) Toolchain {
	return &release{
		version: version,
		os:      os,
		arch:    arch,
	}
}

func (r *release) String() string {
	return path.Join("release", r.version, r.os, r.arch)
}

func (r *release) Install(w *Workspace, root string) {
	defer lg.Scope(w, "release_install")()

	// Log parameters.
	lg.Param(w, "release_version", r.version)
	lg.Param(w, "release_os", r.os)
	lg.Param(w, "release_arch", r.arch)

	// Determine download URL.
	// TODO(mbm): fetch files list in json format
	const base = "https://golang.org/dl/"
	filename := fmt.Sprintf("%s.%s-%s.tar.gz", r.version, r.os, r.arch)
	url := base + filename
	lg.Param(w, "release_url", url)

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
