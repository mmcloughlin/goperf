package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mholt/archiver"
	"github.com/mmcloughlin/cb/pkg/lg"
)

type Workspace struct {
	*log.Logger

	client *http.Client
	root   string
	env    []string
	err    error
}

type Option func(*Workspace)

func WithHTTPClient(c *http.Client) Option {
	return func(w *Workspace) { w.client = c }
}

func WithLogger(l *log.Logger) Option {
	return func(w *Workspace) { w.Logger = l }
}

func WithWorkDir(d string) Option {
	return func(w *Workspace) { w.root = d }
}

func WithEnviron(env []string) Option {
	return func(w *Workspace) { w.env = append(w.env, env...) }
}

func InheritEnviron() Option {
	return WithEnviron(os.Environ())
}

func NewWorkspace(opts ...Option) (*Workspace, error) {
	// Defaults.
	w := &Workspace{
		Logger: log.New(os.Stderr, "", log.LstdFlags),
		client: http.DefaultClient,
	}

	// Apply options.
	w.Options(opts...)

	// Use a temporary directory if none was specified.
	if w.root == "" {
		dir, err := ioutil.TempDir("", "contbench")
		if err != nil {
			return nil, fmt.Errorf("create working directory: %w", err)
		}
		w.root = dir
	}

	w.Printf("workspace intialized")
	lg.Param(w, "working directory", w.root)

	return w, nil
}

// Options applies options to the workspace.
func (w *Workspace) Options(opts ...Option) {
	for _, opt := range opts {
		opt(w)
	}
}

// Error returns the first error that occurred in the workspace, if any.
func (w *Workspace) Error() error { return w.err }

func (w *Workspace) seterr(err error) {
	if w.err == nil && err != nil {
		w.err = err
	}
}

func (w *Workspace) cancelled() bool {
	return w.err != nil
}

// Clean up the workspace.
func (w *Workspace) Clean() {
	w.seterr(os.RemoveAll(w.root))
}

// SetEnv sets an environment variable for all workspace operations.
func (w *Workspace) SetEnv(key, value string) {
	w.env = append(w.env, key+"="+value)
}

// Path relative to working directory.
func (w *Workspace) Path(rel string) string {
	return filepath.Join(w.root, rel)
}

// EnsureDir ensure the relative path exists.
func (w *Workspace) EnsureDir(rel string) string {
	dir := w.Path(rel)
	if !w.cancelled() {
		w.seterr(os.MkdirAll(dir, 0777))
	}
	return dir
}

// Download url to path.
func (w *Workspace) Download(url, path string) {
	if w.cancelled() {
		return
	}

	defer lg.Scope(w, "download")()
	lg.Param(w, "download_url", url)
	lg.Param(w, "download_path", path)

	// Open file for writing.
	f, err := os.Create(path)
	if err != nil {
		w.seterr(err)
		return
	}
	defer f.Close()

	// Issue request.
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		w.seterr(err)
		return
	}

	res, err := w.client.Do(req)
	if err != nil {
		w.seterr(err)
	}
	defer res.Body.Close()

	// Copy.
	_, err = io.Copy(f, res.Body)
	w.seterr(err)
}

// Uncompress archive src to the directory dst.
func (w *Workspace) Uncompress(src, dst string) {
	if w.cancelled() {
		return
	}
	defer lg.Scope(w, "uncompress")()
	lg.Param(w, "source", src)
	lg.Param(w, "destination", dst)
	w.seterr(archiver.Unarchive(src, dst))
}
