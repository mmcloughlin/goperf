package httputil

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/golang/gddo/httputil"

	"github.com/mmcloughlin/cb/pkg/fs"
	"github.com/mmcloughlin/cb/pkg/lg"
)

// Error is an error with an associated HTTP status code.
type Error struct {
	Code int
	Err  error
}

func (e Error) Error() string {
	var reason string
	if e.Err != nil {
		reason = e.Err.Error()
	} else {
		reason = http.StatusText(e.Code)
	}
	return fmt.Sprintf("http status %d: %s", e.Code, reason)
}

// Status returns the associated HTTP status code.
func (e Error) Status() int { return e.Code }

// InternalServerError builds an error with StatusInternalServerError.
func InternalServerError(err error) Error {
	return Error{Code: http.StatusInternalServerError, Err: err}
}

// NotFound builds a 404 error.
func NotFound() Error {
	return Error{Code: http.StatusNotFound}
}

// BadRequest builds an error with StatusBadRequest.
func BadRequest(err error) Error {
	return Error{Code: http.StatusBadRequest, Err: err}
}

// Handler handles a HTTP request and returns a possible error.
type Handler interface {
	HandleRequest(w http.ResponseWriter, r *http.Request) error
}

// HandlerFunc adapts a function to the Handler interface.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// HandleRequest calls h.
func (h HandlerFunc) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	return h(w, r)
}

// ErrorHandler wraps a HandlerFunc with an error handling layer.
type ErrorHandler struct {
	Handler Handler
	Logger  lg.Logger
}

func (h ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Delegate to the handler, writing to a buffer.
	buf := new(httputil.ResponseBuffer)
	err := h.Handler.HandleRequest(buf, r)

	// On success, copy buffer to the original writer w.
	if err == nil {
		err := buf.WriteTo(w)
		if err != nil {
			lg.Error(h.Logger, "write http response", err)
		}
		return
	}

	// Handle error.
	e, ok := err.(Error)
	if !ok {
		e = InternalServerError(err)
	}

	lg.Error(h.Logger, "handle http request", e)
	http.Error(w, e.Error(), e.Status())
}

// OK responds with an ok response. Intended for serverless handlers.
func OK(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

// NewStatic serves static content from the supplied filesystem.
func NewStatic(filesys fs.Readable) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		name := r.URL.Path

		info, err := filesys.Stat(ctx, name)
		if err != nil {
			return err
		}

		b, err := fs.ReadFile(ctx, filesys, name)
		if err != nil {
			return err
		}

		http.ServeContent(w, r, name, info.ModTime, bytes.NewReader(b))
		return nil
	})
}
