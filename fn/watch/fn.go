package watch

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/repo"
	"github.com/mmcloughlin/cb/app/service"
)

// Services.
var repository = repo.Go(http.DefaultClient)

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Open database connection.
	d, err := service.DB(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer d.Close()

	// Get most recent commit in the database.
	latest, err := d.MostRecentCommit(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("latest commit in database: %s", latest.SHA)

	// Fetch commits until we get to the latest one.
	start := "master"
	for {
		// Fetch commits.
		log.Printf("git log %s", start)
		commits, err := repository.Log(ctx, start)
		if err != nil {
			log.Printf("recent commits: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("returned %d commits", len(commits))

		// Store in database.
		if err := d.StoreCommits(ctx, commits); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("inserted %d commits", len(commits))

		// Look to see if we've hit the latest one.
		if containsCommit(commits, latest) {
			log.Printf("done: found commit %s", latest.SHA)
			break
		}

		// Update log starting point.
		start = commits[len(commits)-1].SHA
	}

	// Report ok.
	fmt.Fprintln(w, "ok")
}

func containsCommit(commits []*entity.Commit, target *entity.Commit) bool {
	for _, c := range commits {
		if c.SHA == target.SHA {
			return true
		}
	}
	return false
}
