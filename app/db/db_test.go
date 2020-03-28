package db

import (
	"context"
	"flag"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcloughlin/cb/app/internal/fixture"
)

var conn = flag.String("conn", "", "database connection string")

// Database opens a database connection.
func Database(t *testing.T) *DB {
	if *conn == "" {
		t.Skip("no database connection string provided")
	}

	db, err := Open(*conn)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	})

	return db
}

func TestDBCommit(t *testing.T) {
	db := Database(t)

	// Store.
	ctx := context.Background()
	err := db.StoreCommit(ctx, fixture.Commit)
	if err != nil {
		t.Fatal(err)
	}

	// Find.
	got, err := db.FindCommitBySHA(ctx, fixture.Commit.SHA)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(fixture.Commit, got); diff != "" {
		t.Errorf("mismatch\n%s", diff)
	}
}

func TestDBModule(t *testing.T) {
	db := Database(t)

	// Store.
	ctx := context.Background()
	expect := fixture.Module
	err := db.StoreModule(ctx, expect)
	if err != nil {
		t.Fatal(err)
	}

	// Find.
	got, err := db.FindModuleByUUID(ctx, expect.UUID())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expect, got); diff != "" {
		t.Errorf("mismatch\n%s", diff)
	}
}

func TestDBPackage(t *testing.T) {
	db := Database(t)

	// Store.
	ctx := context.Background()
	expect := fixture.Package
	err := db.StorePackage(ctx, expect)
	if err != nil {
		t.Fatal(err)
	}

	// Find.
	got, err := db.FindPackageByUUID(ctx, expect.UUID())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expect, got); diff != "" {
		t.Errorf("mismatch\n%s", diff)
	}
}

func TestDBBenchmark(t *testing.T) {
	db := Database(t)

	// Store.
	ctx := context.Background()
	expect := fixture.Benchmark
	err := db.StoreBenchmark(ctx, expect)
	if err != nil {
		t.Fatal(err)
	}

	// Find.
	got, err := db.FindBenchmarkByUUID(ctx, expect.UUID())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expect, got); diff != "" {
		t.Errorf("mismatch\n%s", diff)
	}
}

func TestDBDataFile(t *testing.T) {
	db := Database(t)

	// Store.
	ctx := context.Background()
	expect := fixture.DataFile
	err := db.StoreDataFile(ctx, expect)
	if err != nil {
		t.Fatal(err)
	}

	// Find.
	got, err := db.FindDataFileByUUID(ctx, expect.UUID())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expect, got); diff != "" {
		t.Errorf("mismatch\n%s", diff)
	}
}
