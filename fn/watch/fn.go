package watch

import (
	"net/http"

	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/httputil"
	"github.com/mmcloughlin/cb/app/repo"
	"github.com/mmcloughlin/cb/app/service"
	"github.com/mmcloughlin/cb/pkg/lg"
)

// Services.
var (
	repository = repo.Go(http.DefaultClient)
	logger     = lg.Default()
	handler    = httputil.ErrorHandler{
		Handler: httputil.HandlerFunc(handle),
		Logger:  logger,
	}
)

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}

func handle(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Open database connection.
	d, err := service.DB(ctx)
	if err != nil {
		return err
	}
	defer d.Close()

	// Get most recent commit in the database.
	latest, err := d.MostRecentCommit(ctx)
	if err != nil {
		return err
	}
	logger.Printf("latest commit in database: %s", latest.SHA)

	// Fetch commits until we get to the latest one.
	start := "master"
	for {
		// Fetch commits.
		logger.Printf("git log %s", start)
		commits, err := repository.Log(ctx, start)
		if err != nil {
			logger.Printf("recent commits: %s", err)
			return err
		}

		logger.Printf("returned %d commits", len(commits))

		// Store in database.
		if err := d.StoreCommits(ctx, commits); err != nil {
			return err
		}
		logger.Printf("inserted %d commits", len(commits))

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
