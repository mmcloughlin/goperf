package watch

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mmcloughlin/cb/gitiles"
	"github.com/mmcloughlin/cb/repo"
)

// Parameters.
const (
	gitilesbase = "https://go.googlesource.com"
	gitilesrepo = "go"
)

// Services.
var (
	repolog   repo.Log
	repostore repo.Store
)

// One-time initialization.
func init() {
	// Repository log.
	repolog = repo.NewGitilesLog(
		gitiles.NewClient(http.DefaultClient, gitilesbase),
		gitilesrepo,
	)
}

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch commits.
	commits, err := repolog.RecentCommits(ctx)
	if err != nil {
		log.Printf("recent commits: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, c := range commits {
		log.Print(c.SHA, c.CommitTime)
	}

	// Report ok.
	fmt.Fprintln(w, "ok")
}
