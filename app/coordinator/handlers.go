package coordinator

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/mmcloughlin/cb/app/httputil"
	"github.com/mmcloughlin/cb/pkg/lg"
)

type Handlers struct {
	c *Coordinator

	router  *httprouter.Router
	jsonenc *httputil.JSONEncoder
	logger  lg.Logger
}

func NewHandlers(c *Coordinator, l lg.Logger) *Handlers {
	// Configure.
	h := &Handlers{
		c:       c,
		jsonenc: &httputil.JSONEncoder{Debug: true},
		router:  httprouter.New(),
		logger:  l,
	}

	// Setup mux.
	h.router.Handler(http.MethodPost, "/workers/:worker/jobs", httputil.ErrorHandler{
		Handler: httputil.HandlerFunc(h.RequestJobs),
		Logger:  h.logger,
	})

	return h
}

func (h *Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *Handlers) RequestJobs(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	params := httprouter.ParamsFromContext(r.Context())

	// Build jobs request.
	req := &JobsRequest{
		Worker: params.ByName("worker"),
	}

	// Delegate to Coordinator.
	res, err := h.c.Jobs(ctx, req)
	if err != nil {
		return err
	}

	// Encode response.
	if len(res.Jobs) > 0 {
		w.WriteHeader(http.StatusCreated)
	}
	return h.jsonenc.EncodeResponse(w, res)
}
