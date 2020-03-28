package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/db/internal/db"
	"github.com/mmcloughlin/cb/app/entity"
)

//go:generate rm -rf internal/db
//go:generate sqlc generate

// DB provides a database storage layer.
type DB struct {
	db *sql.DB
	q  *db.Queries
}

// Open postgres database connection with the given connection string.
func Open(conn string) (*DB, error) {
	d, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	return &DB{
		db: d,
		q:  db.New(d),
	}, nil
}

// Close database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// tx executes the given query function in a transaction.
func (d *DB) tx(ctx context.Context, fn func(q *db.Queries) error) (err error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	return fn(d.q.WithTx(tx))
}

// StoreCommit writes commit to the database.
func (d *DB) StoreCommit(ctx context.Context, c *entity.Commit) error {
	return d.tx(ctx, func(q *db.Queries) error {
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

// FindCommitBySHA looks up the given commit in the database.
func (d *DB) FindCommitBySHA(ctx context.Context, sha string) (*entity.Commit, error) {
	shabytes, err := hex.DecodeString(sha)
	if err != nil {
		return nil, fmt.Errorf("invalid sha: %w", err)
	}

	var c *entity.Commit
	err = d.tx(ctx, func(q *db.Queries) error {
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
	}, nil
}

// StoreModule writes module to the database.
func (d *DB) StoreModule(ctx context.Context, m *entity.Module) error {
	return d.tx(ctx, func(q *db.Queries) error {
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
	err := d.tx(ctx, func(q *db.Queries) error {
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
	err := d.tx(ctx, func(q *db.Queries) error {
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
	return d.tx(ctx, func(q *db.Queries) error {
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
	err := d.tx(ctx, func(q *db.Queries) error {
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

	return &entity.Package{
		Module:       m,
		RelativePath: p.RelativePath,
	}, nil
}

// StoreBenchmark writes benchmark to the database.
func (d *DB) StoreBenchmark(ctx context.Context, b *entity.Benchmark) error {
	return d.tx(ctx, func(q *db.Queries) error {
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
	err := d.tx(ctx, func(q *db.Queries) error {
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

	params := map[string]string{}
	if err := json.Unmarshal(b.Parameters, &params); err != nil {
		return nil, fmt.Errorf("decode parameters: %w", err)
	}

	p, err := findPackageByUUID(ctx, q, b.PackageUUID)
	if err != nil {
		return nil, err
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
	return d.tx(ctx, func(q *db.Queries) error {
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
	err := d.tx(ctx, func(q *db.Queries) error {
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
	return d.tx(ctx, func(q *db.Queries) error {
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
	err := d.tx(ctx, func(q *db.Queries) error {
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
	return d.tx(ctx, func(q *db.Queries) error {
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
	err := d.tx(ctx, func(q *db.Queries) error {
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
