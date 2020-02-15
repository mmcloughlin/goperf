package main

import (
	"fmt"
	"path"
	"path/filepath"

	"golang.org/x/build/buildenv"

	"github.com/mmcloughlin/cb/pkg/lg"
)

type Toolchain interface {
	// TODO(mbm): Toolchain returns configuration lines rather than a string
	String() string
	Install(w *Workspace, root string)
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
	filename := fmt.Sprintf("go%s.%s-%s.tar.gz", r.version, r.os, r.arch)
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
