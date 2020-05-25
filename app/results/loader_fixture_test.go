package results_test

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/mod/module"

	"github.com/mmcloughlin/goperf/app/entity"
	"github.com/mmcloughlin/goperf/app/internal/fixture"
	"github.com/mmcloughlin/goperf/app/results"
	"github.com/mmcloughlin/goperf/pkg/cfg"
	"github.com/mmcloughlin/goperf/pkg/fs"
	"github.com/mmcloughlin/goperf/pkg/mod"
)

func TestLoader(t *testing.T) {
	ref := "go1.23.4"

	// Setup configuration lines.
	keys := results.Keys{
		ToolchainRef: "ref",
		Module:       "mod",
		Package:      "pkg",
	}

	c := cfg.Configuration{
		cfg.KeyValue(cfg.Key(keys.ToolchainRef), cfg.StringValue(ref)),
		cfg.KeyValue(cfg.Key(keys.Module), module.Version{Path: fixture.Module.Path, Version: fixture.ModuleSHA}),
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
		Mod:     fixture.Module.Path,
		Rev:     fixture.ModuleSHA,
		RevInfo: fixture.RevInfo,
	}

	// Construct loader.
	loader, err := results.NewLoader(
		results.WithFilesystem(m),
		results.WithRevisions(rev),
		results.WithModuleInfo(moddb),
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

func TestLoadSHA256Hash(t *testing.T) {
	// This test purely checks that the loader correctly computes the SHA-256 hash of the file.
	//
	//	$ sha256sum testdata/e5a4b8be-c0e5-42c4-a243-2b458ceff483.txt
	//	a36064ebcadaea9b4b419ef66b487a6bdf1f0d5f90efa513d35f800d4dfceeb1  testdata/e5a4b8be-c0e5-42c4-a243-2b458ceff483.txt

	filename := "e5a4b8be-c0e5-42c4-a243-2b458ceff483.txt"
	expect := "a36064ebcadaea9b4b419ef66b487a6bdf1f0d5f90efa513d35f800d4dfceeb1"

	fs := fs.NewLocal("testdata")
	loader, err := results.NewLoader(
		results.WithFilesystem(fs),
		results.WithRevisions(&SingleRevision{Commit: fixture.Commit}),
		results.WithModuleInfo(&SingleModule{RevInfo: fixture.RevInfo}),
	)
	if err != nil {
		t.Fatal(err)
	}

	rs, err := loader.Load(context.Background(), filename)
	if err != nil {
		t.Fatal(err)
	}

	if len(rs) == 0 {
		t.Fatal("no results")
	}

	for _, r := range rs {
		got := hex.EncodeToString(r.File.SHA256[:])
		if got != expect {
			t.Fatalf("got hash %s; expect %s", got, expect)
		}
	}
}

// Revision is an implementation of repo.Revisions that returns a fixed commit.
type SingleRevision struct {
	Ref    string
	Commit *entity.Commit
}

func (r *SingleRevision) Revision(_ context.Context, ref string) (*entity.Commit, error) {
	if r.Ref != "" && ref != r.Ref {
		return nil, errors.New("unknown")
	}
	return r.Commit, nil
}

// SingleModule is an implementation of mod.Infoer that returns fixed revision
// information.
type SingleModule struct {
	Mod, Rev string
	RevInfo  *mod.RevInfo
}

func (m *SingleModule) Info(_ context.Context, mod, rev string) (*mod.RevInfo, error) {
	if (m.Mod != "" && mod != m.Mod) || (m.Rev != "" && rev != m.Rev) {
		return nil, errors.New("unknown")
	}
	return m.RevInfo, nil
}
