package httputil

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"
	"time"
)

type proxysingleurl struct {
	u *url.URL
	c *http.Client

	mu      sync.Mutex
	data    []byte
	modtime time.Time
}

// ProxySingleURL is a handler that serves the content at the given URL.
func ProxySingleURL(u *url.URL) Handler {
	return &proxysingleurl{
		u: u,
		c: http.DefaultClient,
	}
}

func (p *proxysingleurl) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	if r.URL.Path != p.u.Path {
		return NotFound()
	}

	// Fetch if necessary.
	if !p.fetched() {
		if err := p.fetch(r.Context()); err != nil {
			return err
		}
	}

	// Serve.
	http.ServeContent(w, r, filepath.Base(p.u.Path), p.modtime, bytes.NewReader(p.data))
	return nil
}

func (p *proxysingleurl) fetch(ctx context.Context) error {
	// Make HTTP request.
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.u.String(), nil)
	if err != nil {
		return err
	}

	res, err := p.c.Do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// Extract last modified date from headers, if possible.
	modtime := time.Now()
	if t, err := http.ParseTime(res.Header.Get("Last-Modified")); err == nil {
		modtime = t
	}

	// Set.
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data = data
	p.modtime = modtime

	return nil
}

func (p *proxysingleurl) fetched() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.data != nil
}
