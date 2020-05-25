package noop

import (
	"net/http"

	"github.com/mmcloughlin/goperf/app/httputil"
)

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	httputil.OK(w)
}
