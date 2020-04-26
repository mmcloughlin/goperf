// Package db provides a database storage layer.
package db

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/app/db/internal/db"
	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/trace"
	"github.com/mmcloughlin/cb/internal/errutil"
)

//go:generate rm -rf internal/db
//go:generate sqlc generate

// DB provides database access.
type DB struct {
	db  *sql.DB
	q   *db.Queries
	log *zap.Logger
}

// New builds a database layer backed by the given postgres connection.
func New(ctx context.Context, d *sql.DB) (*DB, error) {
	q, err := db.Prepare(ctx, d)
	if err != nil {
		return nil, err
	}
	return &DB{
		db:  d,
		q:   q,
		log: zap.NewNop(),
	}, nil
}

// Open postgres database connection with the given connection string.
func Open(ctx context.Context, conn string) (*DB, error) {
	d, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	return New(ctx, d)
}

// Close database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// SetLogger configures a logger.
func (d *DB) SetLogger(l *zap.Logger) { d.log = l.Named("db") }

// tx executes the given function in a transaction.
func (d *DB) tx(ctx context.Context, fn func(*sql.Tx) error) (err error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		switch p := recover(); {
		case p != nil:
			if err := tx.Rollback(); err != nil {
				d.log.Error("transaction rollback error", zap.Error(err))
			}
			panic(p)
		case err != nil:
			if err := tx.Rollback(); err != nil {
				d.log.Error("transaction rollback error", zap.Error(err))
			}
		default:
			err = tx.Commit()
		}
	}()
	return fn(tx)
}

// txq executes the given query function in a transaction.
func (d *DB) txq(ctx context.Context, fn func(q *db.Queries) error) error {
	return d.tx(ctx, func(tx *sql.Tx) error { return fn(d.q.WithTx(tx)) })
}

// insert executes a batch insert.
func (d *DB) insert(ctx context.Context, tx *sql.Tx, table string, fields []string, values []interface{}, trailer string) error {
	n := len(values)
	if n%len(fields) != 0 {
		return errutil.AssertionFailure("number of values must be a multiple of the number of fields")
	}
	// Build query.
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "INSERT INTO %s (%s) VALUES", table, strings.Join(fields, ","))
	sep := byte(' ')
	for i := 0; i < n; i += len(fields) {
		buf.WriteByte(sep)
		sep = '('
		for j := 0; j < len(fields); j++ {
			buf.WriteByte(sep)
			sep = ','
			buf.WriteByte('$')
			buf.WriteString(strconv.Itoa(i + j + 1))
		}
		buf.WriteByte(')')
		sep = ','
	}
	buf.WriteString(" " + trailer)
	q := buf.String()
	// Execute.
	_, err := tx.ExecContext(ctx, q, values...)
	return err
}

// StoreCommit writes commit to the database.
func (d *DB) StoreCommit(ctx context.Context, c *entity.Commit) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return storeCommit(ctx, q, c)
	})
}

func storeCommit(ctx context.Context, q *db.Queries, c *entity.Commit) error {
	sha, err := hex.DecodeString(c.SHA)
	if err != nil {
		return fmt.Errorf("invalid sha: %w", err)
	}

	tree, err := hex.DecodeString(c.Tree)
	if err != nil {
		return fmt.Errorf("invalid tree: %w", err)
	}

	parents := make([][]byte, len(c.Parents))
	for i, p := range c.Parents {
		parents[i], err = hex.DecodeString(p)
		if err != nil {
			return fmt.Errorf("invalid parent: %w", err)
		}
	}

	return q.InsertCommit(ctx, db.InsertCommitParams{
		SHA:            sha,
		Tree:           tree,
		Parents:        parents,
		AuthorName:     c.Author.Name,
		AuthorEmail:    c.Author.Email,
		AuthorTime:     c.AuthorTime,
		CommitterName:  c.Committer.Name,
		CommitterEmail: c.Committer.Email,
		CommitTime:     c.CommitTime,
		Message:        c.Message,
	})
}

// StoreCommits writes the given commits to the database in a single batch.
func (d *DB) StoreCommits(ctx context.Context, cs []*entity.Commit) error {
	fields := []string{
		"sha",
		"tree",
		"parents",
		"author_name",
		"author_email",
		"author_time",
		"committer_name",
		"committer_email",
		"commit_time",
		"message",
	}
	values := []interface{}{}
	for _, c := range cs {
		sha, err := hex.DecodeString(c.SHA)
		if err != nil {
			return fmt.Errorf("invalid sha: %w", err)
		}

		tree, err := hex.DecodeString(c.Tree)
		if err != nil {
			return fmt.Errorf("invalid tree: %w", err)
		}

		parents := make([][]byte, len(c.Parents))
		for i, p := range c.Parents {
			parents[i], err = hex.DecodeString(p)
			if err != nil {
				return fmt.Errorf("invalid parent: %w", err)
			}
		}

		values = append(values,
			sha,
			tree,
			pq.ByteaArray(parents),
			c.Author.Name,
			c.Author.Email,
			c.AuthorTime,
			c.Committer.Name,
			c.Committer.Email,
			c.CommitTime,
			c.Message,
		)
	}
	return d.tx(ctx, func(tx *sql.Tx) error {
		return d.insert(ctx, tx, "commits", fields, values, "ON CONFLICT DO NOTHING")
	})
}

// FindCommitBySHA looks up the given commit in the database.
func (d *DB) FindCommitBySHA(ctx context.Context, sha string) (*entity.Commit, error) {
	shabytes, err := hex.DecodeString(sha)
	if err != nil {
		return nil, fmt.Errorf("invalid sha: %w", err)
	}

	var c *entity.Commit
	err = d.txq(ctx, func(q *db.Queries) error {
		var err error
		c, err = findCommitBySHA(ctx, q, shabytes)
		return err
	})
	return c, err
}

func findCommitBySHA(ctx context.Context, q *db.Queries, sha []byte) (*entity.Commit, error) {
	c, err := q.Commit(ctx, sha)
	if err != nil {
		return nil, err
	}

	return mapCommit(c), nil
}

// MostRecentCommit returns the most recent commit by commit time.
func (d *DB) MostRecentCommit(ctx context.Context) (*entity.Commit, error) {
	var c *entity.Commit
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		c, err = mostRecentCommit(ctx, q)
		return err
	})
	return c, err
}

func mostRecentCommit(ctx context.Context, q *db.Queries) (*entity.Commit, error) {
	c, err := q.MostRecentCommit(ctx)
	if err != nil {
		return nil, err
	}

	return mapCommit(c), nil
}

// MostRecentCommitWithRef returns the most recent commit by commit time having the supplied ref.
func (d *DB) MostRecentCommitWithRef(ctx context.Context, ref string) (*entity.Commit, error) {
	var c *entity.Commit
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		c, err = mostRecentCommitWithRef(ctx, q, ref)
		return err
	})
	return c, err
}

func mostRecentCommitWithRef(ctx context.Context, q *db.Queries, ref string) (*entity.Commit, error) {
	c, err := q.MostRecentCommitWithRef(ctx, ref)
	if err != nil {
		return nil, err
	}

	return mapCommit(c), nil
}

func mapCommit(c db.Commit) *entity.Commit {
	parents := make([]string, len(c.Parents))
	for i, parent := range c.Parents {
		parents[i] = hex.EncodeToString(parent)
	}

	return &entity.Commit{
		SHA:     hex.EncodeToString(c.SHA),
		Tree:    hex.EncodeToString(c.Tree),
		Parents: parents,
		Author: entity.Person{
			Name:  c.AuthorName,
			Email: c.AuthorEmail,
		},
		AuthorTime: c.AuthorTime,
		Committer: entity.Person{
			Name:  c.CommitterName,
			Email: c.CommitterEmail,
		},
		CommitTime: c.CommitTime,
		Message:    c.Message,
	}
}

// StoreCommitRef writes a commit ref pair to the database.
func (d *DB) StoreCommitRef(ctx context.Context, r *entity.CommitRef) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return storeCommitRef(ctx, q, r)
	})
}

// StoreCommitRefs writes the given commit refs to the database.
func (d *DB) StoreCommitRefs(ctx context.Context, rs []*entity.CommitRef) error {
	fields := []string{
		"sha",
		"ref",
	}
	values := []interface{}{}
	for _, r := range rs {
		sha, err := hex.DecodeString(r.SHA)
		if err != nil {
			return fmt.Errorf("invalid sha: %w", err)
		}

		values = append(values,
			sha,
			r.Ref,
		)
	}
	return d.tx(ctx, func(tx *sql.Tx) error {
		return d.insert(ctx, tx, "commit_refs", fields, values, "ON CONFLICT DO NOTHING")
	})
}

func storeCommitRef(ctx context.Context, q *db.Queries, r *entity.CommitRef) error {
	sha, err := hex.DecodeString(r.SHA)
	if err != nil {
		return fmt.Errorf("invalid sha: %w", err)
	}

	return q.InsertCommitRef(ctx, db.InsertCommitRefParams{
		SHA: sha,
		Ref: r.Ref,
	})
}

// BuildCommitPositions creates the commit positions table. The table is
// completely rebuilt from the source tables.
func (d *DB) BuildCommitPositions(ctx context.Context) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return q.BuildCommitPositions(ctx)
	})
}

// MostRecentCommitIndex returns the most recent commit index.
func (d *DB) MostRecentCommitIndex(ctx context.Context) (int, error) {
	var idx int
	err := d.txq(ctx, func(q *db.Queries) error {
		i, err := q.MostRecentCommitIndex(ctx)
		idx = int(i)
		return err
	})
	return idx, err
}

// StoreModule writes module to the database.
func (d *DB) StoreModule(ctx context.Context, m *entity.Module) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return storeModule(ctx, q, m)
	})
}

func storeModule(ctx context.Context, q *db.Queries, m *entity.Module) error {
	return q.InsertModule(ctx, db.InsertModuleParams{
		UUID:    m.UUID(),
		Path:    m.Path,
		Version: m.Version,
	})
}

// FindModuleByUUID looks up the given module in the database.
func (d *DB) FindModuleByUUID(ctx context.Context, id uuid.UUID) (*entity.Module, error) {
	var m *entity.Module
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		m, err = findModuleByUUID(ctx, q, id)
		return err
	})
	return m, err
}

func findModuleByUUID(ctx context.Context, q *db.Queries, id uuid.UUID) (*entity.Module, error) {
	m, err := q.Module(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapModule(m), nil
}

// ListModules returns all modules.
func (d *DB) ListModules(ctx context.Context) ([]*entity.Module, error) {
	var ms []*entity.Module
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		ms, err = listModules(ctx, q)
		return err
	})
	return ms, err
}

func listModules(ctx context.Context, q *db.Queries) ([]*entity.Module, error) {
	ms, err := q.Modules(ctx)
	if err != nil {
		return nil, err
	}

	output := make([]*entity.Module, len(ms))
	for i, m := range ms {
		output[i] = mapModule(m)
	}

	return output, nil
}

func mapModule(m db.Module) *entity.Module {
	return &entity.Module{
		Path:    m.Path,
		Version: m.Version,
	}
}

// StorePackage writes package to the database.
func (d *DB) StorePackage(ctx context.Context, p *entity.Package) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return storePackage(ctx, q, p)
	})
}

func storePackage(ctx context.Context, q *db.Queries, p *entity.Package) error {
	if err := storeModule(ctx, q, p.Module); err != nil {
		return err
	}

	return q.InsertPkg(ctx, db.InsertPkgParams{
		UUID:         p.UUID(),
		ModuleUUID:   p.Module.UUID(),
		RelativePath: p.RelativePath,
	})
}

// FindPackageByUUID looks up the given package in the database.
func (d *DB) FindPackageByUUID(ctx context.Context, id uuid.UUID) (*entity.Package, error) {
	var p *entity.Package
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		p, err = findPackageByUUID(ctx, q, id)
		return err
	})
	return p, err
}

func findPackageByUUID(ctx context.Context, q *db.Queries, id uuid.UUID) (*entity.Package, error) {
	p, err := q.Pkg(ctx, id)
	if err != nil {
		return nil, err
	}

	m, err := findModuleByUUID(ctx, q, p.ModuleUUID)
	if err != nil {
		return nil, err
	}

	return mapPackage(p, m), nil
}

// ListModulePackages returns all packages in the given module.
func (d *DB) ListModulePackages(ctx context.Context, m *entity.Module) ([]*entity.Package, error) {
	var ps []*entity.Package
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		ps, err = listModulePackages(ctx, q, m)
		return err
	})
	return ps, err
}

func listModulePackages(ctx context.Context, q *db.Queries, m *entity.Module) ([]*entity.Package, error) {
	ps, err := q.ModulePkgs(ctx, m.UUID())
	if err != nil {
		return nil, err
	}

	output := make([]*entity.Package, len(ps))
	for i, p := range ps {
		output[i] = mapPackage(p, m)
	}

	return output, nil
}

func mapPackage(p db.Package, m *entity.Module) *entity.Package {
	return &entity.Package{
		Module:       m,
		RelativePath: p.RelativePath,
	}
}

// StoreBenchmark writes benchmark to the database.
func (d *DB) StoreBenchmark(ctx context.Context, b *entity.Benchmark) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return storeBenchmark(ctx, q, b)
	})
}

func storeBenchmark(ctx context.Context, q *db.Queries, b *entity.Benchmark) error {
	if err := storePackage(ctx, q, b.Package); err != nil {
		return err
	}

	paramsjson, err := json.Marshal(b.Parameters)
	if err != nil {
		return fmt.Errorf("encode parameters: %w", err)
	}

	return q.InsertBenchmark(ctx, db.InsertBenchmarkParams{
		UUID:        b.UUID(),
		PackageUUID: b.Package.UUID(),
		FullName:    b.FullName,
		Name:        b.Name,
		Unit:        b.Unit,
		Parameters:  paramsjson,
	})
}

// FindBenchmarkByUUID looks up the given benchmark in the database.
func (d *DB) FindBenchmarkByUUID(ctx context.Context, id uuid.UUID) (*entity.Benchmark, error) {
	var b *entity.Benchmark
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		b, err = findBenchmarkByUUID(ctx, q, id)
		return err
	})
	return b, err
}

func findBenchmarkByUUID(ctx context.Context, q *db.Queries, id uuid.UUID) (*entity.Benchmark, error) {
	b, err := q.Benchmark(ctx, id)
	if err != nil {
		return nil, err
	}

	p, err := findPackageByUUID(ctx, q, b.PackageUUID)
	if err != nil {
		return nil, err
	}

	return mapBenchmark(b, p)
}

// ListPackageBenchmarks returns all benchmarks in the given package.
func (d *DB) ListPackageBenchmarks(ctx context.Context, p *entity.Package) ([]*entity.Benchmark, error) {
	var bs []*entity.Benchmark
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		bs, err = listPackageBenchmarks(ctx, q, p)
		return err
	})
	return bs, err
}

func listPackageBenchmarks(ctx context.Context, q *db.Queries, p *entity.Package) ([]*entity.Benchmark, error) {
	bs, err := q.PackageBenchmarks(ctx, p.UUID())
	if err != nil {
		return nil, err
	}

	output := make([]*entity.Benchmark, len(bs))
	for i, b := range bs {
		output[i], err = mapBenchmark(b, p)
		if err != nil {
			return nil, err
		}
	}

	return output, nil
}

func mapBenchmark(b db.Benchmark, p *entity.Package) (*entity.Benchmark, error) {
	params := map[string]string{}
	if err := json.Unmarshal(b.Parameters, &params); err != nil {
		return nil, fmt.Errorf("decode parameters: %w", err)
	}

	return &entity.Benchmark{
		Package:    p,
		FullName:   b.FullName,
		Name:       b.Name,
		Unit:       b.Unit,
		Parameters: params,
	}, nil
}

// StoreDataFile writes the data file to the database.
func (d *DB) StoreDataFile(ctx context.Context, f *entity.DataFile) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return storeDataFile(ctx, q, f)
	})
}

func storeDataFile(ctx context.Context, q *db.Queries, f *entity.DataFile) error {
	return q.InsertDataFile(ctx, db.InsertDataFileParams{
		UUID:   f.UUID(),
		Name:   f.Name,
		SHA256: f.SHA256[:],
	})
}

// FindDataFileByUUID looks up the given data file in the database.
func (d *DB) FindDataFileByUUID(ctx context.Context, id uuid.UUID) (*entity.DataFile, error) {
	var f *entity.DataFile
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		f, err = findDataFileByUUID(ctx, q, id)
		return err
	})
	return f, err
}

func findDataFileByUUID(ctx context.Context, q *db.Queries, id uuid.UUID) (*entity.DataFile, error) {
	f, err := q.DataFile(ctx, id)
	if err != nil {
		return nil, err
	}

	var hash [sha256.Size]byte
	copy(hash[:], f.SHA256)

	return &entity.DataFile{
		Name:   f.Name,
		SHA256: hash,
	}, nil
}

// StoreProperties writes properties to the database.
func (d *DB) StoreProperties(ctx context.Context, p entity.Properties) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return storeProperties(ctx, q, p)
	})
}

func storeProperties(ctx context.Context, q *db.Queries, p entity.Properties) error {
	propertiesjson, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("encode properties: %w", err)
	}

	return q.InsertProperties(ctx, db.InsertPropertiesParams{
		UUID:   p.UUID(),
		Fields: propertiesjson,
	})
}

// FindPropertiesByUUID looks up the given properties in the database.
func (d *DB) FindPropertiesByUUID(ctx context.Context, id uuid.UUID) (entity.Properties, error) {
	var p entity.Properties
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		p, err = findPropertiesByUUID(ctx, q, id)
		return err
	})
	return p, err
}

func findPropertiesByUUID(ctx context.Context, q *db.Queries, id uuid.UUID) (entity.Properties, error) {
	p, err := q.Properties(ctx, id)
	if err != nil {
		return nil, err
	}

	var properties entity.Properties
	if err := json.Unmarshal(p.Fields, &properties); err != nil {
		return nil, fmt.Errorf("decode properties: %w", err)
	}

	return properties, nil
}

// StoreResult writes a result to the database.
func (d *DB) StoreResult(ctx context.Context, r *entity.Result) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return storeResult(ctx, q, r)
	})
}

func storeResult(ctx context.Context, q *db.Queries, r *entity.Result) error {
	// DataFile.
	if err := storeDataFile(ctx, q, r.File); err != nil {
		return err
	}

	// Benchmark.
	if err := storeBenchmark(ctx, q, r.Benchmark); err != nil {
		return err
	}

	// Commit.
	if err := storeCommit(ctx, q, r.Commit); err != nil {
		return err
	}

	sha, err := hex.DecodeString(r.Commit.SHA)
	if err != nil {
		return fmt.Errorf("invalid sha: %w", err)
	}

	// Environment and metadata.
	if err := storeProperties(ctx, q, r.Environment); err != nil {
		return err
	}

	if err := storeProperties(ctx, q, r.Metadata); err != nil {
		return err
	}

	return q.InsertResult(ctx, db.InsertResultParams{
		UUID:            r.UUID(),
		DatafileUUID:    r.File.UUID(),
		Line:            int32(r.Line),
		BenchmarkUUID:   r.Benchmark.UUID(),
		CommitSHA:       sha,
		EnvironmentUUID: r.Environment.UUID(),
		MetadataUUID:    r.Metadata.UUID(),
		Iterations:      int64(r.Iterations),
		Value:           r.Value,
	})
}

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

// ListBenchmarkPoints returns timeseries points for the given benchmark and commit time range.
func (d *DB) ListBenchmarkPoints(ctx context.Context, b *entity.Benchmark, ref string, start, end time.Time) (entity.Points, error) {
	var ps []*entity.Point
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		ps, err = listBenchmarkPoints(ctx, q, b, ref, start, end)
		return err
	})
	return ps, err
}

func listBenchmarkPoints(ctx context.Context, q *db.Queries, b *entity.Benchmark, ref string, start, end time.Time) (entity.Points, error) {
	ps, err := q.BenchmarkPoints(ctx, db.BenchmarkPointsParams{
		Ref:             ref,
		BenchmarkUUID:   b.UUID(),
		CommitTimeStart: start,
		CommitTimeEnd:   end,
	})
	if err != nil {
		return nil, err
	}

	// Convert to point objects. Reverse at the same time, since the query returns in descending order.
	output := make(entity.Points, len(ps))
	for i, p := range ps {
		j := len(output) - i - 1
		output[j] = &entity.Point{
			ResultUUID:      p.ResultUUID,
			EnvironmentUUID: p.EnvironmentUUID,
			CommitSHA:       hex.EncodeToString(p.CommitSHA),
			CommitTime:      p.CommitTime,
			Value:           p.Value,
		}
	}

	return output, nil
}

// ListTracePoints returns trace points for the given commit range.
func (d *DB) ListTracePoints(ctx context.Context, r entity.CommitIndexRange) ([]trace.Point, error) {
	var ps []trace.Point
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		ps, err = listTracePoints(ctx, q, r)
		return err
	})
	return ps, err
}

func listTracePoints(ctx context.Context, q *db.Queries, r entity.CommitIndexRange) ([]trace.Point, error) {
	ps, err := q.TracePoints(ctx, db.TracePointsParams{
		CommitIndexMin: int32(r.Min),
		CommitIndexMax: int32(r.Max),
	})
	if err != nil {
		return nil, err
	}

	// Convert to trace point objects.
	output := make([]trace.Point, len(ps))
	for i, p := range ps {
		output[i] = trace.Point{
			BenchmarkUUID:   p.BenchmarkUUID,
			EnvironmentUUID: p.EnvironmentUUID,
			CommitIndex:     int(p.CommitIndex),
			CommitTime:      p.CommitTime,
			Value:           p.Value,
		}
	}

	return output, nil
}
