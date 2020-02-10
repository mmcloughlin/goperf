package noop

import (
	"fmt"
	"net/http"
)

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	// Report ok.
	fmt.Fprintln(w, "ok")
}
