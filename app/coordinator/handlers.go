package coordinator

import (
	"net/http"

	"github.com/mmcloughlin/cb/app/httputil"
)

type Handlers struct {
	c *Coordinator

	mux     *http.ServeMux
	jsondec *httputil.JSONDecoder
	jsonenc *httputil.JSONEncoder
}

func NewHandlers(c *Coordinator) *Handlers {
	// Configure.
	h := &Handlers{
		c:       c,
		jsondec: &httputil.JSONDecoder{MaxRequestSize: 1 << 20},
		mux:     http.NewServeMux(),
	}

	// Setup mux.
	h.mux.HandleFunc("/jobs", h.Jobs)

	return h
}

func (h *Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handlers) Jobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Decode request.
	req := &JobsRequest{}
	if err := h.jsondec.DecodeRequest(w, r, req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Delegate to Coordinator.
	res, err := h.c.Jobs(ctx, req)
	if err != nil {
		httputil.InternalServerError(w, err)
		return
	}

	// Encode response.
	h.jsonenc.EncodeResponse(w, res)
}
