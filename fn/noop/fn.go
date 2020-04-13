package noop

import (
	"net/http"

	"github.com/mmcloughlin/cb/app/httputil"
)

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	httputil.OK(w)
}
