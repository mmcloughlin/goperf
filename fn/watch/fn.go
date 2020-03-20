package watch

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"

	"github.com/mmcloughlin/cb/app/mapper"
	"github.com/mmcloughlin/cb/app/obj"
	"github.com/mmcloughlin/cb/app/repo"
)

// Parameters.
var project = os.Getenv("CB_PROJECT_ID")

// Services.
var repository = repo.Go(http.DefaultClient)

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch commits.
	commits, err := repository.RecentCommits(ctx)
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
	s := obj.NewFirestore(fsc)
	for _, c := range commits {
		m := mapper.CommitToModel(c)
		if err := s.Set(ctx, m); err != nil {
			log.Printf("upsert commit %s: %s", c.SHA, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("upserted commit %s", c.SHA)
	}

	// Report ok.
	fmt.Fprintln(w, "ok")
}
