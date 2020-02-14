package main

import (
	"path"
	"path/filepath"

	"github.com/mmcloughlin/cb/pkg/lg"
	"golang.org/x/build/buildenv"
)

type Toolchain interface {
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
	dldir := w.EnsureDir("dl")
	archive := filepath.Join(dldir, "go.tar.gz")
	w.Download(url, archive)

	// Extract.
	w.Uncompress(archive, root)
}
