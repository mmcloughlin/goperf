package httputil

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CacheControl specifies Cache-Control header options.
type CacheControl struct {
	MaxAge       time.Duration
	SharedMaxAge time.Duration
	Directives   []string
}

// CacheControlImmutable provides suitable cache control headers for immutable files.
var CacheControlImmutable = CacheControl{
	MaxAge:       24 * time.Hour,
	SharedMaxAge: 365 * 24 * time.Hour,
	Directives:   []string{"public", "immutable"},
}

// CacheControlNever defines cache control headers to ensure a resource is never cached.
var CacheControlNever = CacheControl{
	MaxAge:       0,
	SharedMaxAge: 0,
	Directives:   []string{"no-cache", "no-store", "no-transform", "must-revalidate", "private"},
}

func (c CacheControl) String() string {
	directives := c.Directives
	directives = append(directives, fmt.Sprintf("max-age=%d", c.MaxAge/time.Second))
	if c.SharedMaxAge != 0 {
		directives = append(directives, fmt.Sprintf("s-maxage=%d", c.SharedMaxAge/time.Second))
	}
	return strings.Join(directives, ", ")
}

// CacheHandler wraps the handler and sets Cache-Control headers.
func CacheHandler(cc CacheControl, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := w.Header()
		if hdr.Get("Cache-Control") != "" {
			return
		}
		hdr.Set("Cache-Control", cc.String())

		h.ServeHTTP(w, r)
	})
}
