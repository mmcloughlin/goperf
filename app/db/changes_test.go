package db_test

import (
	"context"
	"testing"

	"github.com/mmcloughlin/cb/app/db/dbtest"
	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/internal/fixture"
)

func TestDBStoreChangesBatch(t *testing.T) {
	db := dbtest.Open(t)

	// Ensure the dependenent objects exist.
	ctx := context.Background()
	err := db.StoreBenchmark(ctx, fixture.Change.Benchmark)
	if err != nil {
		t.Fatal(err)
	}

	if err = db.StoreCommit(ctx, fixture.Commit); err != nil {
		t.Fatal(err)
	}

	if err = db.StoreCommitPosition(ctx, fixture.CommitPosition); err != nil {
		t.Fatal(err)
	}

	// Store change in batch mode.
	err = db.StoreChangesBatch(ctx, []*entity.Change{fixture.Change})
	if err != nil {
		t.Fatal(err)
	}
}
