package watch

import (
	"log"
	"net/http"

	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/httputil"
	"github.com/mmcloughlin/cb/app/repo"
	"github.com/mmcloughlin/cb/app/service"
)

// Services.
var (
	repository repo.Repository
	logger     *zap.Logger
	handler    http.Handler
)

func init() {
	var err error

	repository = repo.Go(http.DefaultClient)

	logger, err = service.Logger()
	if err != nil {
		log.Fatal(err)
	}

	handler = httputil.ErrorHandler{
		Handler: httputil.HandlerFunc(handle),
		Log:     logger,
	}
}

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}

func handle(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Open database connection.
	d, err := service.DB(ctx, logger)
	if err != nil {
		return err
	}
	defer d.Close()

	// Get most recent commit in the database.
	latest, err := d.MostRecentCommit(ctx)
	if err != nil {
		return err
	}
	logger.Info("found latest commit in database", zap.String("sha", latest.SHA))

	// Fetch commits until we get to the latest one.
	start := "master"
	for {
		// Fetch commits.
		logger.Info("git log", zap.String("start", start))
		commits, err := repository.Log(ctx, start)
		if err != nil {
			logger.Error("error fetching recent commits", zap.Error(err))
			return err
		}

		logger.Info("fetched recent commits", zap.Int("num_commits", len(commits)))

		// Store in database.
		if err := d.StoreCommits(ctx, commits); err != nil {
			return err
		}
		logger.Info("inserted commits", zap.Int("num_commits", len(commits)))

		// Look to see if we've hit the latest one.
		if containsCommit(commits, latest) {
			break
		}

		// Update log starting point.
		start = commits[len(commits)-1].SHA
	}

	// Report ok.
	httputil.OK(w)

	return nil
}

func containsCommit(commits []*entity.Commit, target *entity.Commit) bool {
	for _, c := range commits {
		if c.SHA == target.SHA {
			return true
		}
	}
	return false
}
