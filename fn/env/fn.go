package env

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	// Log environment.
	for _, e := range os.Environ() {
		log.Println(e)
	}

	// Report ok.
	fmt.Fprintln(w, "ok")
}
