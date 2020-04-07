package coordinator

import (
	"net/http"

	"github.com/mmcloughlin/cb/app/httputil"
	"github.com/mmcloughlin/cb/pkg/lg"
)

type Handlers struct {
	c *Coordinator

	mux     *http.ServeMux
	jsondec *httputil.JSONDecoder
	jsonenc *httputil.JSONEncoder
	logger  lg.Logger
}

func NewHandlers(c *Coordinator, l lg.Logger) *Handlers {
	// Configure.
	h := &Handlers{
		c:       c,
		jsondec: &httputil.JSONDecoder{MaxRequestSize: 1 << 20},
		mux:     http.NewServeMux(),
		logger:  l,
	}

	// Setup mux.
	h.mux.Handle("/jobs", httputil.ErrorHandler{
		Handler: httputil.HandlerFunc(h.Jobs),
		Logger:  h.logger,
	})

	return h
}

func (h *Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handlers) Jobs(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Decode request.
	req := &JobsRequest{}
	if err := h.jsondec.DecodeRequest(w, r, req); err != nil {
		return httputil.BadRequest(err)
	}

	// Delegate to Coordinator.
	res, err := h.c.Jobs(ctx, req)
	if err != nil {
		return err
	}

	// Encode response.
	return h.jsonenc.EncodeResponse(w, res)
}
