package watch

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/app/change"
	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/httputil"
	"github.com/mmcloughlin/cb/app/service"
	"github.com/mmcloughlin/cb/app/trace"
)

// NumCommits is the number of most recent commits to look for changes in.
const NumCommits = 512

// Initialization.
var (
	logger   *zap.Logger
	database *db.DB
	handler  http.Handler
	detector = change.DefaultDetector
)

func initialize(ctx context.Context, l *zap.Logger) error {
	var err error

	logger = l

	database, err = service.DB(ctx, l)
	if err != nil {
		return err
	}

	handler = httputil.ErrorHandler{
		Handler: httputil.HandlerFunc(handle),
		Log:     l,
	}

	return nil
}

func init() {
	service.Initialize(initialize)
}

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}

func handle(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Determine commit range.
	idx, err := database.MostRecentCommitIndex(ctx)
	if err != nil {
		return err
	}
	logger.Info("most recent commit index", zap.Int("index", idx))

	cr := entity.CommitIndexRange{
		Min: idx - NumCommits + 1,
		Max: idx,
	}

	// Query for trace points.
	logger.Info("fetching traces",
		zap.Int("min_commit_index", cr.Min),
		zap.Int("max_commit_index", cr.Max),
	)

	ps, err := database.ListTracePoints(ctx, cr)
	if err != nil {
		return err
	}

	traces := trace.Traces(ps)

	// Find change points.
	var changes []*entity.Change
	for id, trc := range traces {
		log := logger.With(zap.Stringer("trace", id))

		chgs := detector.Detect(trc.Series)
		if len(chgs) == 0 {
			continue
		}

		for _, chg := range chgs {
			log.Info("change found",
				zap.Int("commit_index", chg.CommitIndex),
				zap.Float64("effect_size", chg.EffectSize),
			)
			changes = append(changes, &entity.Change{
				ID:     id,
				Change: chg,
			})
		}
	}

	// Insert into database.
	if err := database.ReplaceChanges(ctx, cr, changes); err != nil {
		return err
	}

	// Report ok.
	httputil.OK(w)

	return nil
}
