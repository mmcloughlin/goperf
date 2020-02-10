package watch

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"

	"github.com/mmcloughlin/cb/gitiles"
	"github.com/mmcloughlin/cb/repo"
)

// Parameters.
var (
	project = os.Getenv("CB_PROJECT_ID")
)

const (
	gitilesbase = "https://go.googlesource.com"
	gitilesrepo = "go"

	commitscollection = "commits"
)

// Services.
var (
	repolog repo.Log
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

	// Create Firestore client.
	fsc, err := firestore.NewClient(ctx, project)
	if err != nil {
		log.Printf("firestore new client: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Write commits to store.
	s := repo.NewFirestoreStore(fsc, commitscollection)
	for _, c := range commits {
		if err := s.Upsert(ctx, c); err != nil {
			log.Printf("upsert commit %s: %s", c.SHA, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("upserted commit %s", c.SHA)
	}

	// Report ok.
	fmt.Fprintln(w, "ok")
}
