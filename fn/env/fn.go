package env

import (
	"log"
	"net/http"
	"os"

	"github.com/mmcloughlin/cb/app/httputil"
)

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	// Log environment.
	for _, e := range os.Environ() {
		log.Println(e)
	}

	// Report ok.
	httputil.OK(w)
}
