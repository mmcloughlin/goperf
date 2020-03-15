package results_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcloughlin/cb/app/internal/fixture"
	"github.com/mmcloughlin/cb/app/repo"
	"github.com/mmcloughlin/cb/app/results"
	"github.com/mmcloughlin/cb/app/suite"
	"github.com/mmcloughlin/cb/pkg/cfg"
	"github.com/mmcloughlin/cb/pkg/fs"
)

func TestLoader(t *testing.T) {
	var ref = "go1.23.4"

	// Setup configuration lines.
	keys := results.Keys{
		ToolchainRef:  "ref",
		ModulePath:    "modpath",
		ModuleVersion: "modversion",
		Package:       "pkg",
	}

	c := cfg.Configuration{
		cfg.KeyValue(cfg.Key(keys.ToolchainRef), cfg.StringValue(ref)),
		cfg.KeyValue(cfg.Key(keys.ModulePath), cfg.StringValue(fixture.Module.Path)),
		cfg.KeyValue(cfg.Key(keys.ModuleVersion), cfg.StringValue(fixture.ModuleSHA)),
		cfg.KeyValue(cfg.Key(keys.Package), cfg.StringValue(fixture.Package.ImportPath())),
	}

	envtags := []cfg.Tag{"perf0", "perf1"}
	for k, v := range fixture.Result.Environment {
		tag := envtags[len(c)%len(envtags)]
		c = append(c, cfg.KeyValue(cfg.Key(k), cfg.StringValue(v), tag))
	}

	for k, v := range fixture.Result.Metadata {
		c = append(c, cfg.KeyValue(cfg.Key(k), cfg.StringValue(v)))
	}

	// Write a datafile to an in-memory filesystem.
	ctx := context.Background()
	m := fs.NewMem()
	w, err := m.Create(ctx, fixture.DataFile.Name)
	if err != nil {
		t.Fatal(err)
	}

	if err := cfg.Write(w, c); err != nil {
		t.Fatal(err)
	}

	r := fixture.Result
	_, err = fmt.Fprintln(w, r.Benchmark.FullName, r.Iterations, r.Value, r.Benchmark.Unit)
	if err != nil {
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	// Read it back and log.
	b, err := fs.ReadFile(ctx, m, fixture.DataFile.Name)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("datafile:\n%s", b)

	// Setup mock revision and module providers.
	rev := &SingleRevision{
		Ref:    ref,
		Commit: fixture.Commit,
	}

	moddb := &SingleModule{
		Mod:  fixture.Module.Path,
		Rev:  fixture.ModuleSHA,
		Info: fixture.RevInfo,
	}

	// Construct loader.
	loader, err := results.NewLoader(
		results.WithFilesystem(m),
		results.WithRevisions(rev),
		results.WithModuleDatabase(moddb),
		results.WithKeys(keys),
		results.WithEnvironmentTags(envtags...),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Load results.
	rs, err := loader.Load(ctx, fixture.DataFile.Name)
	if err != nil {
		t.Fatal(err)
	}

	if len(rs) != 1 {
		t.Fatalf("got %d results; expect one", len(rs))
	}
	r = rs[0]

	// Compare against expectation. Handle line number and SHA specially.
	if r.Line != len(c)+1 {
		t.Errorf("got line number %d; expect %d", r.Line, len(c)+1)
	}

	r.Line = fixture.Result.Line
	r.File.SHA256 = fixture.DataFile.SHA256
	if diff := cmp.Diff(fixture.Result, r); diff != "" {
		t.Errorf("mismatch\n%s", diff)
	}
}

// Revision is an implementation of repo.Revisions that returns a fixed commit.
type SingleRevision struct {
	Ref    string
	Commit *repo.Commit
}

func (r *SingleRevision) Revision(_ context.Context, ref string) (*repo.Commit, error) {
	if ref != r.Ref {
		return nil, errors.New("unknown")
	}
	return r.Commit, nil
}

// SingleModule is an implementation of suite.ModuleDatabase that returns fixed
// revision information.
type SingleModule struct {
	Mod, Rev string
	Info     *suite.RevInfo
}

func (m *SingleModule) Stat(_ context.Context, mod, rev string) (*suite.RevInfo, error) {
	if mod != m.Mod || rev != m.Rev {
		return nil, errors.New("unknown")
	}
	return m.Info, nil
}
