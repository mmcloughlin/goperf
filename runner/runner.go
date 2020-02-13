package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
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
	rev   = "v1.10.0"
)

func main() {
	os.Exit(main1())
}

func main1() int {
	if err := mainerr(); err != nil {
		log.Print(err)
		return 1
	}
	return 0
}

func mainerr() error {
	r := NewRunner(buildertype, goversion)

	if err := r.Init(); err != nil {
		return err
	}

	mod := Module{
		Path:    path.Join("github.com", owner, repo),
		Version: rev,
	}
	job := Job{
		Module: mod,
	}
	if err := r.Benchmark(job); err != nil {
		return err
	}

	return nil
}

type Module struct {
	Path    string
	Version string
}

func (m Module) String() string {
	s := m.Path
	if m.Version != "" {
		s += "@" + m.Version
	}
	return s
}

type Job struct {
	Module Module
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

	dldir, err := r.ensuredir("dl")
	if err != nil {
		return err
	}

	archive := filepath.Join(dldir, "go.tar.gz")
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
	gopath, err := r.ensuredir("gopath")
	if err != nil {
		return err
	}
	r.SetEnv("GOPATH", gopath)

	// Environment checks.
	// TODO(mbm): clean these up
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

// Benchmark runs the benchmark job.
func (r *Runner) Benchmark(j Job) error {
	defer r.scope("benchmark")()

	// Get a run directory.
	wd, err := r.rundir("bench")
	if err != nil {
		return err
	}
	r.logparam("working directory", wd)

	// Initialize a module.
	cmd := r.Go("mod", "init", "mod")
	cmd.Dir = wd
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("output:\n%s", out)
		return err
	}

	// Fetch module.
	cmd = r.Go("get", "-t", j.Module.String())
	cmd.Dir = wd
	out, err := cmd.CombinedOutput()
	log.Printf("output:\n%s", out)
	if err != nil {
		return err
	}

	// Run benchmarks.
	cmd = r.Go(
		"test",
		"-run", "none^", // no tests
		"-bench", ".", // all benchmarks
		"-benchtime", "10ms", // 10ms each
		j.Module.Path+"/...",
	)
	cmd.Dir = wd
	out, err = cmd.CombinedOutput()
	log.Printf("output:\n%s", out)
	if err != nil {
		return err
	}

	return nil
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

// rundir generates a new working directory for a run.
func (r *Runner) rundir(prefix string) (string, error) {
	rundir, err := r.ensuredir("run")
	if err != nil {
		return "", err
	}
	return ioutil.TempDir(rundir, prefix)
}

// ensuredir ensure the relative path exists.
func (r *Runner) ensuredir(rel string) (string, error) {
	dir := r.path(rel)
	if err := os.MkdirAll(r.path(rel), 0777); err != nil {
		return "", err
	}
	return dir, nil
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
