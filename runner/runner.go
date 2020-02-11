package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/mholt/archiver"

	"golang.org/x/build/buildenv"
)

const (
	goversion   = "60d437f99468906935f35e5c6fbd31c7228a1045"
	buildertype = "darwin-amd64-10_14"

	owner = "klauspost"
	repo  = "compress"
	rev   = "b7ccab840e50d2c3fbfdb00c8949cc32e31cd459"
)

func main() {
	r := NewRunner(buildertype, goversion)
	if err := r.Init(); err != nil {
		log.Print(err)
	}
}

type Runner struct {
	buildertype string
	goversion   string

	client *http.Client
	logger *log.Logger

	workdir string
	env     []string
	gobin   string
}

type Option func(*Runner)

func WithHTTPClient(c *http.Client) Option {
	return func(r *Runner) { r.client = c }
}

func WithLogger(l *log.Logger) Option {
	return func(r *Runner) { r.logger = l }
}

func WithWorkDir(d string) Option {
	return func(r *Runner) { r.workdir = d }
}

func NewRunner(buildertype, goversion string, opts ...Option) *Runner {
	r := &Runner{
		buildertype: buildertype,
		goversion:   goversion,
		client:      http.DefaultClient,
		logger:      log.New(os.Stderr, "", log.LstdFlags),
	}
	r.Options(opts...)
	return r
}

func (r *Runner) Options(opts ...Option) {
	for _, opt := range opts {
		opt(r)
	}
}

// Init initializes the runner.
func (r *Runner) Init() error {
	defer r.scope("initializing")()

	// Determine a working directory.
	if r.workdir == "" {
		dir, err := ioutil.TempDir("", "contbench")
		if err != nil {
			return fmt.Errorf("create working directory: %w", err)
		}
		r.workdir = dir
	}
	r.logparam("working directory", r.workdir)

	// Download go release.
	r.logparam("builder type", r.buildertype)
	r.logparam("go version", r.goversion)
	url := buildenv.Production.SnapshotURL(r.buildertype, r.goversion)
	r.logparam("snapshot url", url)

	if err := r.ensuredir("dl"); err != nil {
		return err
	}

	archive := r.path("dl/go.tar.gz")
	if err := r.download(url, archive); err != nil {
		return err
	}

	goroot := r.path("goroot")
	if err := r.uncompress(archive, goroot); err != nil {
		return err
	}
	r.SetEnv("GOROOT", goroot)

	r.gobin = filepath.Join(goroot, "bin", "go")

	// Configure GOPATH.
	gopath := r.path("gopath")
	if err := r.ensuredir("gopath"); err != nil {
		return err
	}
	r.SetEnv("GOPATH", gopath)

	// Environment checks.
	cmd := r.Go("version")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = r.Go("env")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// SetEnv sets an environment variable for all runner operations.
func (r *Runner) SetEnv(key, value string) {
	r.env = append(r.env, key+"="+value)
}

// Close cleans up after the runner.
func (r *Runner) Close() error {
	return os.RemoveAll(r.workdir)
}

// Go runs a command with the downloaded go version.
func (r *Runner) Go(arg ...string) *exec.Cmd {
	return &exec.Cmd{
		Path: r.gobin,
		Args: append([]string{"go"}, arg...),
		Env:  append(os.Environ(), r.env...),
	}
}

func (r *Runner) download(url, path string) error {
	defer r.scope("download")()
	r.logparam("download url", url)
	r.logparam("download path", path)

	// Open file for writing.
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Issue request.
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Copy.
	_, err = io.Copy(f, res.Body)
	return err
}

func (r *Runner) uncompress(src, dst string) error {
	defer r.scope("uncompress")()
	r.logparam("source", src)
	r.logparam("destination", dst)
	return archiver.Unarchive(src, dst)
}

// ensuredir ensure the relative path exists.
func (r *Runner) ensuredir(rel string) error {
	return os.MkdirAll(r.path(rel), 0777)
}

// path relative to working directory.
func (r *Runner) path(rel string) string {
	return filepath.Join(r.workdir, rel)
}

func (r *Runner) logparam(key, value string) {
	r.logger.Printf("%s = %s\n", key, value)
}

func (r *Runner) scope(name string) func() {
	t0 := time.Now()
	r.logger.Printf("start: %s", name)
	return func() {
		r.logger.Printf("finish: %s (time %s)", name, time.Since(t0))
	}
}
