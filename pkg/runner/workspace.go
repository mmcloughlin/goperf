package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/pkg/fs"
	"github.com/mmcloughlin/goperf/pkg/lg"
)

type Workspace struct {
	Log *zap.Logger

	client    *http.Client
	artifacts fs.Writable

	root string
	cwd  string
	env  map[string]string
	err  error
}

type Option func(*Workspace)

func WithHTTPClient(c *http.Client) Option {
	return func(w *Workspace) { w.client = c }
}

func WithLogger(l *zap.Logger) Option {
	return func(w *Workspace) { w.Log = l }
}

func WithWorkDir(d string) Option {
	return func(w *Workspace) { w.root = d }
}

func WithEnviron(env []string) Option {
	return func(w *Workspace) { w.AddEnviron(env...) }
}

func InheritEnviron() Option {
	return WithEnviron(os.Environ())
}

func WithArtifactStore(fs fs.Writable) Option {
	return func(w *Workspace) { w.artifacts = fs }
}

func NewWorkspace(opts ...Option) (*Workspace, error) {
	// Defaults.
	w := &Workspace{
		Log:       zap.NewNop(),
		client:    http.DefaultClient,
		artifacts: fs.Null,
		env:       map[string]string{},
	}

	// Apply options.
	w.Options(opts...)

	// Use a temporary directory if none was specified.
	if w.root == "" {
		dir, err := ioutil.TempDir("", "bench")
		if err != nil {
			return nil, fmt.Errorf("create working directory: %w", err)
		}
		w.root = dir
	}

	w.Log.Info("workspace intialized")

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

func (w *Workspace) cancelled() bool {
	return w.err != nil
}

func (w *Workspace) seterr(err error) {
	if w.err == nil && err != nil {
		w.err = err
	}
}

func (w *Workspace) close(c io.Closer) {
	w.seterr(c.Close())
}

// Clean up the workspace.
func (w *Workspace) Clean() {
	w.seterr(os.RemoveAll(w.root))
}

// SetEnv sets an environment variable for all workspace operations.
func (w *Workspace) SetEnv(key, value string) {
	w.Log.Info("set environment variable", zap.String(key, value))
	w.env[key] = value
}

// SetEnvDefault sets an environment variable if it does not already have a value.
func (w *Workspace) SetEnvDefault(key, value string) string {
	if existing := w.GetEnv(key); existing == "" {
		w.SetEnv(key, value)
	}
	return w.GetEnv(key)
}

// AddEnviron is a convenience for setting multiple environment variables given
// a list of "KEY=value" strings. Provided for easy interoperability with
// functions like os.Environ().
func (w *Workspace) AddEnviron(env ...string) {
	for _, e := range env {
		kv := strings.SplitN(e, "=", 2)
		if len(kv) != 2 {
			w.seterr(fmt.Errorf("invalid environment variable setting %q", e))
		}
		w.SetEnv(kv[0], kv[1])
	}
}

// InheritEnv sets the environment variable key to the same as the surrounding
// environment, if it is defined. Otherwise does nothing.
func (w *Workspace) InheritEnv(key string) {
	if v := os.Getenv(key); v != "" {
		w.SetEnv(key, v)
	}
}

// GetEnv returns the environment variable key.
func (w *Workspace) GetEnv(key string) string {
	return w.env[key]
}

// environ returns the configured environment as a list of "KEY=value" strings.
func (w *Workspace) environ() []string {
	var env []string
	for k, v := range w.env {
		env = append(env, k+"="+v)
	}
	return env
}

// AppendPATH appends a directory to the PATH variable, if it is not already present.
func (w *Workspace) AppendPATH(path string) {
	paths := filepath.SplitList(w.GetEnv("PATH"))
	for _, p := range paths {
		if p == path {
			return
		}
	}
	paths = append(paths, path)
	w.SetEnv("PATH", strings.Join(paths, string(filepath.ListSeparator)))
}

// ExposeTool makes the named tool available to the workspace by looking up its
// location and adding the directory to the PATH.
func (w *Workspace) ExposeTool(name string) {
	path, err := exec.LookPath(name)
	if err != nil {
		w.seterr(err)
	}
	w.AppendPATH(filepath.Dir(path))
}

// DefineTool defines a standard tool with environment variable key and default
// dflt, for example "CC" with default "gcc". If the environment variable is set
// in the host environment, it is inherited, otherwise it is set to the default
// and the PATH is edited to ensure it is accessible within the workspace.
func (w *Workspace) DefineTool(key, dflt string) {
	w.InheritEnv(key)
	name := w.SetEnvDefault(key, dflt)
	w.ExposeTool(name)
}

// Path relative to working directory.
func (w *Workspace) Path(rel string) string {
	return filepath.Join(w.root, rel)
}

// EnsureDir ensure the relative path exists.
func (w *Workspace) EnsureDir(rel string) string {
	dir := w.Path(rel)
	if !w.cancelled() {
		w.seterr(os.MkdirAll(dir, 0o777))
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

	defer lg.Scope(w.Log, "download",
		zap.String("url", url),
		zap.String("path", path),
	)()

	// Open file for writing.
	f, err := os.Create(path)
	if err != nil {
		w.seterr(err)
		return
	}
	defer w.close(f)

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
	defer w.close(res.Body)

	// Copy.
	_, err = io.Copy(f, res.Body)
	w.seterr(err)
}

// Uncompress archive src to the directory dst.
func (w *Workspace) Uncompress(src, dst string) {
	if w.cancelled() {
		return
	}
	defer lg.Scope(w.Log, "uncompress",
		zap.String("source", src),
		zap.String("destination", dst),
	)()
	w.seterr(archiver.Unarchive(src, dst))
}

// Move src to dst.
func (w *Workspace) Move(src, dst string) {
	if w.cancelled() {
		return
	}

	defer lg.Scope(w.Log, "move",
		zap.String("source", src),
		zap.String("destination", dst),
	)()

	w.seterr(os.Rename(src, dst))
}

// Cd sets the working directory to path.
func (w *Workspace) Cd(path string) {
	w.cwd = path
	w.Log.Debug("change working directory", zap.String("cwd", w.cwd))
}

// CdRoot sets the working directory to the root of the workspace.
func (w *Workspace) CdRoot() { w.Cd(w.root) }

// Exec the provided command.
func (w *Workspace) Exec(cmd *exec.Cmd) {
	if w.cancelled() {
		return
	}

	defer lg.Scope(w.Log, "exec")()

	// Set environment.
	cmd.Env = append(cmd.Env, w.environ()...)

	// Set working directory.
	if cmd.Dir == "" {
		cmd.Dir = w.cwd
	}

	// Capture output.
	var stdout, stderr bytes.Buffer
	cmd.Stdout = tee(cmd.Stdout, &stdout)
	cmd.Stderr = tee(cmd.Stderr, &stderr)

	w.Log.Info("command prepared", zap.Stringer("cmd", cmd))
	err := cmd.Run()

	w.Log.Info("command complete",
		zap.ByteString("stdout", stdout.Bytes()),
		zap.ByteString("stderr", stderr.Bytes()),
	)

	w.seterr(err)
}

func tee(w, t io.Writer) io.Writer {
	if w == nil {
		return t
	}
	return io.MultiWriter(w, t)
}

// Artifact saves the given path as a named artifact.
func (w *Workspace) Artifact(path, name string) {
	if w.cancelled() {
		return
	}

	defer lg.Scope(w.Log, "artifact",
		zap.String("source", path),
		zap.String("name", name),
	)()

	// Open file to be saved.
	src, err := os.Open(path)
	if err != nil {
		w.seterr(err)
		return
	}
	defer w.close(src)

	// Create destination.
	dst, err := w.artifacts.Create(context.TODO(), name)
	if err != nil {
		w.seterr(err)
		return
	}
	defer w.close(dst)

	// Copy.
	_, err = io.Copy(dst, src)
	w.seterr(err)
}
