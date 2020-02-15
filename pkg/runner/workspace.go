package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mholt/archiver"

	"github.com/mmcloughlin/cb/pkg/lg"
)

type Workspace struct {
	lg.Logger

	client *http.Client
	root   string
	cwd    string
	env    []string
	err    error
}

type Option func(*Workspace)

func WithHTTPClient(c *http.Client) Option {
	return func(w *Workspace) { w.client = c }
}

func WithLogger(l lg.Logger) Option {
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
		Logger: lg.Default(),
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

	// Start working directory at root.
	w.CdRoot()

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

// Sandbox creates a fresh temporary directory, sets it as the working directory
// and returns it.
func (w *Workspace) Sandbox(task string) string {
	if w.cancelled() {
		return ""
	}
	sandbox := w.EnsureDir("sandbox")
	dir, err := ioutil.TempDir(sandbox, task)
	w.seterr(err)
	w.Cd(dir)
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

// Move src to dst.
func (w *Workspace) Move(src, dst string) {
	if w.cancelled() {
		return
	}

	defer lg.Scope(w, "move")()
	lg.Param(w, "source", src)
	lg.Param(w, "destination", dst)

	w.seterr(os.Rename(src, dst))
}

// Cd sets the working directory to path.
func (w *Workspace) Cd(path string) {
	w.cwd = path
	lg.Param(w, "working directory", w.cwd)
}

// CdRoot sets the working directory to the root of the workspace.
func (w *Workspace) CdRoot() { w.Cd(w.root) }

// Exec the provided command.
func (w *Workspace) Exec(cmd *exec.Cmd) {
	if w.cancelled() {
		return
	}

	defer lg.Scope(w, "exec")()

	// Set environment.
	cmd.Env = append(cmd.Env, w.env...)

	// Set working directory.
	if cmd.Dir == "" {
		cmd.Dir = w.cwd
	}

	// Capture output.
	var stdout, stderr bytes.Buffer
	cmd.Stdout = tee(cmd.Stdout, &stdout)
	cmd.Stderr = tee(cmd.Stderr, &stderr)

	lg.Param(w, "cmd", cmd)
	err := cmd.Run()

	lg.Param(w, "stdout", stdout.String())
	lg.Param(w, "stderr", stderr.String())

	w.seterr(err)
}

func tee(w, t io.Writer) io.Writer {
	if w == nil {
		return t
	}
	return io.MultiWriter(w, t)
}
