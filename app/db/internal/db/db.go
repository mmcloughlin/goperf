// Code generated by sqlc. DO NOT EDIT.

package db

import (
	"context"
	"database/sql"
	"fmt"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

func Prepare(ctx context.Context, db DBTX) (*Queries, error) {
	q := Queries{db: db}
	var err error
	if q.benchmarkStmt, err = db.PrepareContext(ctx, benchmark); err != nil {
		return nil, fmt.Errorf("error preparing query Benchmark: %w", err)
	}
	if q.benchmarkPointsStmt, err = db.PrepareContext(ctx, benchmarkPoints); err != nil {
		return nil, fmt.Errorf("error preparing query BenchmarkPoints: %w", err)
	}
	if q.benchmarkResultsStmt, err = db.PrepareContext(ctx, benchmarkResults); err != nil {
		return nil, fmt.Errorf("error preparing query BenchmarkResults: %w", err)
	}
	if q.buildCommitPositionsStmt, err = db.PrepareContext(ctx, buildCommitPositions); err != nil {
		return nil, fmt.Errorf("error preparing query BuildCommitPositions: %w", err)
	}
	if q.commitStmt, err = db.PrepareContext(ctx, commit); err != nil {
		return nil, fmt.Errorf("error preparing query Commit: %w", err)
	}
	if q.commitModuleWorkerErrorsStmt, err = db.PrepareContext(ctx, commitModuleWorkerErrors); err != nil {
		return nil, fmt.Errorf("error preparing query CommitModuleWorkerErrors: %w", err)
	}
	if q.createTaskStmt, err = db.PrepareContext(ctx, createTask); err != nil {
		return nil, fmt.Errorf("error preparing query CreateTask: %w", err)
	}
	if q.dataFileStmt, err = db.PrepareContext(ctx, dataFile); err != nil {
		return nil, fmt.Errorf("error preparing query DataFile: %w", err)
	}
	if q.deleteChangesCommitRangeStmt, err = db.PrepareContext(ctx, deleteChangesCommitRange); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteChangesCommitRange: %w", err)
	}
	if q.insertBenchmarkStmt, err = db.PrepareContext(ctx, insertBenchmark); err != nil {
		return nil, fmt.Errorf("error preparing query InsertBenchmark: %w", err)
	}
	if q.insertCommitStmt, err = db.PrepareContext(ctx, insertCommit); err != nil {
		return nil, fmt.Errorf("error preparing query InsertCommit: %w", err)
	}
	if q.insertCommitPositionStmt, err = db.PrepareContext(ctx, insertCommitPosition); err != nil {
		return nil, fmt.Errorf("error preparing query InsertCommitPosition: %w", err)
	}
	if q.insertCommitRefStmt, err = db.PrepareContext(ctx, insertCommitRef); err != nil {
		return nil, fmt.Errorf("error preparing query InsertCommitRef: %w", err)
	}
	if q.insertDataFileStmt, err = db.PrepareContext(ctx, insertDataFile); err != nil {
		return nil, fmt.Errorf("error preparing query InsertDataFile: %w", err)
	}
	if q.insertModuleStmt, err = db.PrepareContext(ctx, insertModule); err != nil {
		return nil, fmt.Errorf("error preparing query InsertModule: %w", err)
	}
	if q.insertPkgStmt, err = db.PrepareContext(ctx, insertPkg); err != nil {
		return nil, fmt.Errorf("error preparing query InsertPkg: %w", err)
	}
	if q.insertPropertiesStmt, err = db.PrepareContext(ctx, insertProperties); err != nil {
		return nil, fmt.Errorf("error preparing query InsertProperties: %w", err)
	}
	if q.insertResultStmt, err = db.PrepareContext(ctx, insertResult); err != nil {
		return nil, fmt.Errorf("error preparing query InsertResult: %w", err)
	}
	if q.moduleStmt, err = db.PrepareContext(ctx, module); err != nil {
		return nil, fmt.Errorf("error preparing query Module: %w", err)
	}
	if q.modulePkgsStmt, err = db.PrepareContext(ctx, modulePkgs); err != nil {
		return nil, fmt.Errorf("error preparing query ModulePkgs: %w", err)
	}
	if q.modulesStmt, err = db.PrepareContext(ctx, modules); err != nil {
		return nil, fmt.Errorf("error preparing query Modules: %w", err)
	}
	if q.mostRecentCommitStmt, err = db.PrepareContext(ctx, mostRecentCommit); err != nil {
		return nil, fmt.Errorf("error preparing query MostRecentCommit: %w", err)
	}
	if q.mostRecentCommitIndexStmt, err = db.PrepareContext(ctx, mostRecentCommitIndex); err != nil {
		return nil, fmt.Errorf("error preparing query MostRecentCommitIndex: %w", err)
	}
	if q.mostRecentCommitWithRefStmt, err = db.PrepareContext(ctx, mostRecentCommitWithRef); err != nil {
		return nil, fmt.Errorf("error preparing query MostRecentCommitWithRef: %w", err)
	}
	if q.packageBenchmarksStmt, err = db.PrepareContext(ctx, packageBenchmarks); err != nil {
		return nil, fmt.Errorf("error preparing query PackageBenchmarks: %w", err)
	}
	if q.pkgStmt, err = db.PrepareContext(ctx, pkg); err != nil {
		return nil, fmt.Errorf("error preparing query Pkg: %w", err)
	}
	if q.propertiesStmt, err = db.PrepareContext(ctx, properties); err != nil {
		return nil, fmt.Errorf("error preparing query Properties: %w", err)
	}
	if q.recentCommitModulePairsWithoutWorkerTasksStmt, err = db.PrepareContext(ctx, recentCommitModulePairsWithoutWorkerTasks); err != nil {
		return nil, fmt.Errorf("error preparing query RecentCommitModulePairsWithoutWorkerTasks: %w", err)
	}
	if q.resultStmt, err = db.PrepareContext(ctx, result); err != nil {
		return nil, fmt.Errorf("error preparing query Result: %w", err)
	}
	if q.setTaskDataFileStmt, err = db.PrepareContext(ctx, setTaskDataFile); err != nil {
		return nil, fmt.Errorf("error preparing query SetTaskDataFile: %w", err)
	}
	if q.taskStmt, err = db.PrepareContext(ctx, task); err != nil {
		return nil, fmt.Errorf("error preparing query Task: %w", err)
	}
	if q.tasksWithStatusStmt, err = db.PrepareContext(ctx, tasksWithStatus); err != nil {
		return nil, fmt.Errorf("error preparing query TasksWithStatus: %w", err)
	}
	if q.tracePointsStmt, err = db.PrepareContext(ctx, tracePoints); err != nil {
		return nil, fmt.Errorf("error preparing query TracePoints: %w", err)
	}
	if q.transitionTaskStatusStmt, err = db.PrepareContext(ctx, transitionTaskStatus); err != nil {
		return nil, fmt.Errorf("error preparing query TransitionTaskStatus: %w", err)
	}
	if q.transitionTaskStatusesBeforeStmt, err = db.PrepareContext(ctx, transitionTaskStatusesBefore); err != nil {
		return nil, fmt.Errorf("error preparing query TransitionTaskStatusesBefore: %w", err)
	}
	if q.truncateNonStaticStmt, err = db.PrepareContext(ctx, truncateNonStatic); err != nil {
		return nil, fmt.Errorf("error preparing query TruncateNonStatic: %w", err)
	}
	if q.workerTasksWithStatusStmt, err = db.PrepareContext(ctx, workerTasksWithStatus); err != nil {
		return nil, fmt.Errorf("error preparing query WorkerTasksWithStatus: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	if q.benchmarkStmt != nil {
		if cerr := q.benchmarkStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing benchmarkStmt: %w", cerr)
		}
	}
	if q.benchmarkPointsStmt != nil {
		if cerr := q.benchmarkPointsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing benchmarkPointsStmt: %w", cerr)
		}
	}
	if q.benchmarkResultsStmt != nil {
		if cerr := q.benchmarkResultsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing benchmarkResultsStmt: %w", cerr)
		}
	}
	if q.buildCommitPositionsStmt != nil {
		if cerr := q.buildCommitPositionsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing buildCommitPositionsStmt: %w", cerr)
		}
	}
	if q.commitStmt != nil {
		if cerr := q.commitStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing commitStmt: %w", cerr)
		}
	}
	if q.commitModuleWorkerErrorsStmt != nil {
		if cerr := q.commitModuleWorkerErrorsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing commitModuleWorkerErrorsStmt: %w", cerr)
		}
	}
	if q.createTaskStmt != nil {
		if cerr := q.createTaskStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createTaskStmt: %w", cerr)
		}
	}
	if q.dataFileStmt != nil {
		if cerr := q.dataFileStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing dataFileStmt: %w", cerr)
		}
	}
	if q.deleteChangesCommitRangeStmt != nil {
		if cerr := q.deleteChangesCommitRangeStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteChangesCommitRangeStmt: %w", cerr)
		}
	}
	if q.insertBenchmarkStmt != nil {
		if cerr := q.insertBenchmarkStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertBenchmarkStmt: %w", cerr)
		}
	}
	if q.insertCommitStmt != nil {
		if cerr := q.insertCommitStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertCommitStmt: %w", cerr)
		}
	}
	if q.insertCommitPositionStmt != nil {
		if cerr := q.insertCommitPositionStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertCommitPositionStmt: %w", cerr)
		}
	}
	if q.insertCommitRefStmt != nil {
		if cerr := q.insertCommitRefStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertCommitRefStmt: %w", cerr)
		}
	}
	if q.insertDataFileStmt != nil {
		if cerr := q.insertDataFileStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertDataFileStmt: %w", cerr)
		}
	}
	if q.insertModuleStmt != nil {
		if cerr := q.insertModuleStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertModuleStmt: %w", cerr)
		}
	}
	if q.insertPkgStmt != nil {
		if cerr := q.insertPkgStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertPkgStmt: %w", cerr)
		}
	}
	if q.insertPropertiesStmt != nil {
		if cerr := q.insertPropertiesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertPropertiesStmt: %w", cerr)
		}
	}
	if q.insertResultStmt != nil {
		if cerr := q.insertResultStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertResultStmt: %w", cerr)
		}
	}
	if q.moduleStmt != nil {
		if cerr := q.moduleStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing moduleStmt: %w", cerr)
		}
	}
	if q.modulePkgsStmt != nil {
		if cerr := q.modulePkgsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing modulePkgsStmt: %w", cerr)
		}
	}
	if q.modulesStmt != nil {
		if cerr := q.modulesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing modulesStmt: %w", cerr)
		}
	}
	if q.mostRecentCommitStmt != nil {
		if cerr := q.mostRecentCommitStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing mostRecentCommitStmt: %w", cerr)
		}
	}
	if q.mostRecentCommitIndexStmt != nil {
		if cerr := q.mostRecentCommitIndexStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing mostRecentCommitIndexStmt: %w", cerr)
		}
	}
	if q.mostRecentCommitWithRefStmt != nil {
		if cerr := q.mostRecentCommitWithRefStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing mostRecentCommitWithRefStmt: %w", cerr)
		}
	}
	if q.packageBenchmarksStmt != nil {
		if cerr := q.packageBenchmarksStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing packageBenchmarksStmt: %w", cerr)
		}
	}
	if q.pkgStmt != nil {
		if cerr := q.pkgStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing pkgStmt: %w", cerr)
		}
	}
	if q.propertiesStmt != nil {
		if cerr := q.propertiesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing propertiesStmt: %w", cerr)
		}
	}
	if q.recentCommitModulePairsWithoutWorkerTasksStmt != nil {
		if cerr := q.recentCommitModulePairsWithoutWorkerTasksStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing recentCommitModulePairsWithoutWorkerTasksStmt: %w", cerr)
		}
	}
	if q.resultStmt != nil {
		if cerr := q.resultStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing resultStmt: %w", cerr)
		}
	}
	if q.setTaskDataFileStmt != nil {
		if cerr := q.setTaskDataFileStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing setTaskDataFileStmt: %w", cerr)
		}
	}
	if q.taskStmt != nil {
		if cerr := q.taskStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing taskStmt: %w", cerr)
		}
	}
	if q.tasksWithStatusStmt != nil {
		if cerr := q.tasksWithStatusStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing tasksWithStatusStmt: %w", cerr)
		}
	}
	if q.tracePointsStmt != nil {
		if cerr := q.tracePointsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing tracePointsStmt: %w", cerr)
		}
	}
	if q.transitionTaskStatusStmt != nil {
		if cerr := q.transitionTaskStatusStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing transitionTaskStatusStmt: %w", cerr)
		}
	}
	if q.transitionTaskStatusesBeforeStmt != nil {
		if cerr := q.transitionTaskStatusesBeforeStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing transitionTaskStatusesBeforeStmt: %w", cerr)
		}
	}
	if q.truncateNonStaticStmt != nil {
		if cerr := q.truncateNonStaticStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing truncateNonStaticStmt: %w", cerr)
		}
	}
	if q.workerTasksWithStatusStmt != nil {
		if cerr := q.workerTasksWithStatusStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing workerTasksWithStatusStmt: %w", cerr)
		}
	}
	return err
}

func (q *Queries) exec(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
	case stmt != nil:
		return stmt.ExecContext(ctx, args...)
	default:
		return q.db.ExecContext(ctx, query, args...)
	}
}

func (q *Queries) query(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryContext(ctx, args...)
	default:
		return q.db.QueryContext(ctx, query, args...)
	}
}

func (q *Queries) queryRow(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) *sql.Row {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryRowContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryRowContext(ctx, args...)
	default:
		return q.db.QueryRowContext(ctx, query, args...)
	}
}

type Queries struct {
	db                                            DBTX
	tx                                            *sql.Tx
	benchmarkStmt                                 *sql.Stmt
	benchmarkPointsStmt                           *sql.Stmt
	benchmarkResultsStmt                          *sql.Stmt
	buildCommitPositionsStmt                      *sql.Stmt
	commitStmt                                    *sql.Stmt
	commitModuleWorkerErrorsStmt                  *sql.Stmt
	createTaskStmt                                *sql.Stmt
	dataFileStmt                                  *sql.Stmt
	deleteChangesCommitRangeStmt                  *sql.Stmt
	insertBenchmarkStmt                           *sql.Stmt
	insertCommitStmt                              *sql.Stmt
	insertCommitPositionStmt                      *sql.Stmt
	insertCommitRefStmt                           *sql.Stmt
	insertDataFileStmt                            *sql.Stmt
	insertModuleStmt                              *sql.Stmt
	insertPkgStmt                                 *sql.Stmt
	insertPropertiesStmt                          *sql.Stmt
	insertResultStmt                              *sql.Stmt
	moduleStmt                                    *sql.Stmt
	modulePkgsStmt                                *sql.Stmt
	modulesStmt                                   *sql.Stmt
	mostRecentCommitStmt                          *sql.Stmt
	mostRecentCommitIndexStmt                     *sql.Stmt
	mostRecentCommitWithRefStmt                   *sql.Stmt
	packageBenchmarksStmt                         *sql.Stmt
	pkgStmt                                       *sql.Stmt
	propertiesStmt                                *sql.Stmt
	recentCommitModulePairsWithoutWorkerTasksStmt *sql.Stmt
	resultStmt                                    *sql.Stmt
	setTaskDataFileStmt                           *sql.Stmt
	taskStmt                                      *sql.Stmt
	tasksWithStatusStmt                           *sql.Stmt
	tracePointsStmt                               *sql.Stmt
	transitionTaskStatusStmt                      *sql.Stmt
	transitionTaskStatusesBeforeStmt              *sql.Stmt
	truncateNonStaticStmt                         *sql.Stmt
	workerTasksWithStatusStmt                     *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:                           tx,
		tx:                           tx,
		benchmarkStmt:                q.benchmarkStmt,
		benchmarkPointsStmt:          q.benchmarkPointsStmt,
		benchmarkResultsStmt:         q.benchmarkResultsStmt,
		buildCommitPositionsStmt:     q.buildCommitPositionsStmt,
		commitStmt:                   q.commitStmt,
		commitModuleWorkerErrorsStmt: q.commitModuleWorkerErrorsStmt,
		createTaskStmt:               q.createTaskStmt,
		dataFileStmt:                 q.dataFileStmt,
		deleteChangesCommitRangeStmt: q.deleteChangesCommitRangeStmt,
		insertBenchmarkStmt:          q.insertBenchmarkStmt,
		insertCommitStmt:             q.insertCommitStmt,
		insertCommitPositionStmt:     q.insertCommitPositionStmt,
		insertCommitRefStmt:          q.insertCommitRefStmt,
		insertDataFileStmt:           q.insertDataFileStmt,
		insertModuleStmt:             q.insertModuleStmt,
		insertPkgStmt:                q.insertPkgStmt,
		insertPropertiesStmt:         q.insertPropertiesStmt,
		insertResultStmt:             q.insertResultStmt,
		moduleStmt:                   q.moduleStmt,
		modulePkgsStmt:               q.modulePkgsStmt,
		modulesStmt:                  q.modulesStmt,
		mostRecentCommitStmt:         q.mostRecentCommitStmt,
		mostRecentCommitIndexStmt:    q.mostRecentCommitIndexStmt,
		mostRecentCommitWithRefStmt:  q.mostRecentCommitWithRefStmt,
		packageBenchmarksStmt:        q.packageBenchmarksStmt,
		pkgStmt:                      q.pkgStmt,
		propertiesStmt:               q.propertiesStmt,
		recentCommitModulePairsWithoutWorkerTasksStmt: q.recentCommitModulePairsWithoutWorkerTasksStmt,
		resultStmt:                       q.resultStmt,
		setTaskDataFileStmt:              q.setTaskDataFileStmt,
		taskStmt:                         q.taskStmt,
		tasksWithStatusStmt:              q.tasksWithStatusStmt,
		tracePointsStmt:                  q.tracePointsStmt,
		transitionTaskStatusStmt:         q.transitionTaskStatusStmt,
		transitionTaskStatusesBeforeStmt: q.transitionTaskStatusesBeforeStmt,
		truncateNonStaticStmt:            q.truncateNonStaticStmt,
		workerTasksWithStatusStmt:        q.workerTasksWithStatusStmt,
	}
}
