package coordinator

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
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

	// Setup router.
	h.router.Handler(http.MethodPost, "/workers/:worker/jobs", httputil.ErrorHandler{
		Handler: httputil.HandlerFunc(h.requestJobs),
		Logger:  h.logger,
	})

	h.router.Handler(http.MethodPut, "/workers/:worker/jobs/:job/start", httputil.ErrorHandler{
		Handler: httputil.HandlerFunc(h.start),
		Logger:  h.logger,
	})

	return h
}

func (h *Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *Handlers) requestJobs(w http.ResponseWriter, r *http.Request) error {
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

func (h *Handlers) start(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	params := httprouter.ParamsFromContext(r.Context())

	// Build start request.
	id, err := uuid.Parse(params.ByName("job"))
	if err != nil {
		return httputil.BadRequest(fmt.Errorf("bad job uuid: %w", err))
	}

	req := &StartRequest{
		Worker: params.ByName("worker"),
		UUID:   id,
	}

	// Delegate to Coordinator.
	if err := h.c.Start(ctx, req); err != nil {
		return err
	}

	// Return success with no body.
	w.WriteHeader(http.StatusNoContent)
	return nil
}
