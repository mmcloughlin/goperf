package coordinator

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/httputil"
)

type Handlers struct {
	c *Coordinator

	router  *httprouter.Router
	jsonenc *httputil.JSONEncoder
	log     *zap.Logger
}

func NewHandlers(c *Coordinator, l *zap.Logger) *Handlers {
	// Configure.
	h := &Handlers{
		c:       c,
		jsonenc: &httputil.JSONEncoder{Debug: true},
		router:  httprouter.New(),
		log:     l,
	}

	// Setup router.
	h.router.Handler(http.MethodPost, "/workers/:worker/jobs", httputil.ErrorHandler{
		Handler: httputil.HandlerFunc(h.requestJobs),
		Log:     h.log,
	})

	h.router.Handler(http.MethodPut, "/workers/:worker/jobs/:job/start", httputil.ErrorHandler{
		Handler: h.statusChange(
			[]entity.TaskStatus{entity.TaskStatusCreated},
			entity.TaskStatusInProgress,
		),
		Log: h.log,
	})

	h.router.Handler(http.MethodPut, "/workers/:worker/jobs/:job/result", httputil.ErrorHandler{
		Handler: httputil.HandlerFunc(h.result),
		Log:     h.log,
	})

	h.router.Handler(http.MethodPut, "/workers/:worker/jobs/:job/fail", httputil.ErrorHandler{
		Handler: h.statusChange(
			[]entity.TaskStatus{entity.TaskStatusInProgress},
			entity.TaskStatusCompleteError,
		),
		Log: h.log,
	})

	h.router.Handler(http.MethodPut, "/workers/:worker/jobs/:job/halt", httputil.ErrorHandler{
		Handler: h.statusChange(
			entity.TaskStatusPendingValues(),
			entity.TaskStatusHalted,
		),
		Log: h.log,
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

func (h *Handlers) statusChange(from []entity.TaskStatus, to entity.TaskStatus) httputil.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		params := httprouter.ParamsFromContext(r.Context())

		// Build start request.
		id, err := uuid.Parse(params.ByName("job"))
		if err != nil {
			return httputil.BadRequest(fmt.Errorf("bad job uuid: %w", err))
		}

		req := &StatusChangeRequest{
			Worker: params.ByName("worker"),
			UUID:   id,
			From:   from,
			To:     to,
		}

		// Delegate to Coordinator.
		if err := h.c.StatusChange(ctx, req); err != nil {
			return err
		}

		// Return success with no body.
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}

func (h *Handlers) result(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	params := httprouter.ParamsFromContext(r.Context())

	// Build result request.
	id, err := uuid.Parse(params.ByName("job"))
	if err != nil {
		return httputil.BadRequest(fmt.Errorf("bad job uuid: %w", err))
	}

	req := &ResultRequest{
		Reader: r.Body,
		Worker: params.ByName("worker"),
		UUID:   id,
	}

	// Delegate to Coordinator.
	if err := h.c.Result(ctx, req); err != nil {
		return err
	}

	// Return success with no body.
	w.WriteHeader(http.StatusNoContent)
	return nil
}
