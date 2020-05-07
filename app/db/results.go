package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/db/internal/db"
	"github.com/mmcloughlin/cb/app/entity"
)

// FindResultByUUID looks up a result in the database given the ID.
func (d *DB) FindResultByUUID(ctx context.Context, id uuid.UUID) (*entity.Result, error) {
	var r *entity.Result
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		r, err = findResultByUUID(ctx, q, id)
		return err
	})
	return r, err
}

func findResultByUUID(ctx context.Context, q *db.Queries, id uuid.UUID) (*entity.Result, error) {
	r, err := q.Result(ctx, id)
	if err != nil {
		return nil, err
	}

	return result(ctx, q, r)
}

// ListBenchmarkResults returns all results for the given benchmark.
func (d *DB) ListBenchmarkResults(ctx context.Context, b *entity.Benchmark) ([]*entity.Result, error) {
	var rs []*entity.Result
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		rs, err = listBenchmarkResults(ctx, q, b)
		return err
	})
	return rs, err
}

func listBenchmarkResults(ctx context.Context, q *db.Queries, b *entity.Benchmark) ([]*entity.Result, error) {
	rs, err := q.BenchmarkResults(ctx, b.UUID())
	if err != nil {
		return nil, err
	}

	output := make([]*entity.Result, len(rs))
	for i, r := range rs {
		output[i], err = result(ctx, q, r)
		if err != nil {
			return nil, err
		}
	}

	return output, nil
}

func result(ctx context.Context, q *db.Queries, r db.Result) (*entity.Result, error) {
	f, err := findDataFileByUUID(ctx, q, r.DatafileUUID)
	if err != nil {
		return nil, err
	}

	b, err := findBenchmarkByUUID(ctx, q, r.BenchmarkUUID)
	if err != nil {
		return nil, err
	}

	c, err := findCommitBySHA(ctx, q, r.CommitSHA)
	if err != nil {
		return nil, err
	}

	env, err := findPropertiesByUUID(ctx, q, r.EnvironmentUUID)
	if err != nil {
		return nil, err
	}

	meta, err := findPropertiesByUUID(ctx, q, r.MetadataUUID)
	if err != nil {
		return nil, err
	}

	return &entity.Result{
		File:        f,
		Line:        int(r.Line),
		Benchmark:   b,
		Commit:      c,
		Environment: env,
		Metadata:    meta,
		Iterations:  uint64(r.Iterations),
		Value:       r.Value,
	}, nil
}

// StoreResult writes a result to the database.
func (d *DB) StoreResult(ctx context.Context, r *entity.Result) error {
	return d.StoreResults(ctx, []*entity.Result{r})
}

// StoreResults writes results to the database.
func (d *DB) StoreResults(ctx context.Context, rs []*entity.Result) error {
	return d.tx(ctx, func(tx *sql.Tx) error {
		return d.storeResults(ctx, tx, rs)
	})
}

func (d *DB) storeResults(ctx context.Context, tx *sql.Tx, rs []*entity.Result) error {
	q := d.q.WithTx(tx)

	// Construct batch.
	b := newResultBatch()
	b.AddResults(rs)

	// Files.
	for _, f := range b.Files {
		if err := storeDataFile(ctx, q, f); err != nil {
			return err
		}
	}

	// Benchmarks.
	for _, bench := range b.Benchmarks {
		if err := storeBenchmark(ctx, q, bench); err != nil {
			return err
		}
	}

	// Commits.
	for _, c := range b.Commits {
		if err := storeCommit(ctx, q, c); err != nil {
			return err
		}
	}

	// Properties.
	for _, p := range b.Properties {
		if err := storeProperties(ctx, q, p); err != nil {
			return err
		}
	}

	// Results.
	fields := []string{
		"uuid",
		"datafile_uuid",
		"line",
		"benchmark_uuid",
		"commit_sha",
		"environment_uuid",
		"metadata_uuid",
		"iterations",
		"value",
	}
	values := []interface{}{}
	for _, r := range b.Results {
		sha, err := hex.DecodeString(r.Commit.SHA)
		if err != nil {
			return fmt.Errorf("invalid sha: %w", err)
		}
		values = append(values,
			r.UUID(),
			r.File.UUID(),
			int32(r.Line),
			r.Benchmark.UUID(),
			sha,
			r.Environment.UUID(),
			r.Metadata.UUID(),
			int64(r.Iterations),
			r.Value,
		)
	}

	if err := d.insert(ctx, tx, "results", fields, values); err != nil {
		return err
	}

	return nil
}

// resultBatch collects unique objects associated with a set of results.
type resultBatch struct {
	Files      []*entity.DataFile
	Benchmarks []*entity.Benchmark
	Commits    []*entity.Commit
	Properties []entity.Properties
	Results    []*entity.Result

	seen map[uuid.UUID]bool
	shas map[string]bool
}

// newResultBatch builds an empty results batch.
func newResultBatch() *resultBatch {
	return &resultBatch{
		seen: map[uuid.UUID]bool{},
		shas: map[string]bool{},
	}
}

// AddFile adds f to the batch.
func (b *resultBatch) AddFile(f *entity.DataFile) {
	if b.check(f.UUID()) {
		return
	}
	b.Files = append(b.Files, f)
}

// AddBenchmark adds a benchmark to the batch.
func (b *resultBatch) AddBenchmark(bench *entity.Benchmark) {
	if b.check(bench.UUID()) {
		return
	}
	b.Benchmarks = append(b.Benchmarks, bench)
}

// AddCommit adds a commit to the batch.
func (b *resultBatch) AddCommit(c *entity.Commit) {
	if b.shas[c.SHA] {
		return
	}
	b.Commits = append(b.Commits, c)
	b.shas[c.SHA] = true
}

// AddProperties adds a properties set to the batch.
func (b *resultBatch) AddProperties(p entity.Properties) {
	if b.check(p.UUID()) {
		return
	}
	b.Properties = append(b.Properties, p)
}

// AddResult adds a result to the batch.
func (b *resultBatch) AddResult(r *entity.Result) {
	if b.check(r.UUID()) {
		return
	}
	b.AddFile(r.File)
	b.AddBenchmark(r.Benchmark)
	b.AddCommit(r.Commit)
	b.AddProperties(r.Environment)
	b.AddProperties(r.Metadata)
	b.Results = append(b.Results, r)
}

// AddResults adds results to the batch.
func (b *resultBatch) AddResults(rs []*entity.Result) {
	for _, r := range rs {
		b.AddResult(r)
	}
}

// check reports whether the ID is already in the batch, and adds it if not.
func (b *resultBatch) check(id uuid.UUID) bool {
	if b.seen[id] {
		return true
	}
	b.seen[id] = true
	return false
}
