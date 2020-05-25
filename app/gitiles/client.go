// Package gitiles implements a client for the Gitiles API.
package gitiles

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/mmcloughlin/goperf/app/httputil"
	"github.com/mmcloughlin/goperf/internal/errutil"
)

type Client struct {
	client *http.Client
	base   string
}

func NewClient(c *http.Client, base string) *Client {
	return &Client{
		client: c,
		base:   base,
	}
}

func (c *Client) Log(ctx context.Context, repo, ref string) (*LogResponse, error) {
	path := fmt.Sprintf("%s/+log/%s?format=JSON", repo, ref)
	payload := &LogResponse{}
	if err := c.get(ctx, path, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *Client) Revision(ctx context.Context, repo, ref string) (*RevisionResponse, error) {
	path := repo + "/+/" + ref + "?format=JSON"
	payload := &RevisionResponse{}
	if err := c.get(ctx, path, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *Client) get(ctx context.Context, path string, payload interface{}) (err error) {
	// Build request.
	u := c.base + "/" + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}

	// Execute the request.
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer errutil.CheckClose(&err, res.Body)

	// Parse response body.
	if err := decodejson(res.Body, payload); err != nil {
		return err
	}

	return nil
}

func decodejson(rd io.Reader, v interface{}) error {
	r := bufio.NewReader(rd)

	// Expect response to start with "magic" byte sequence.
	magic := []byte(")]}'\n")
	prefix, err := r.Peek(len(magic))
	if err != nil {
		return err
	}
	if !bytes.Equal(magic, prefix) {
		return fmt.Errorf("expected response body to start with magic %q; got %q", magic, prefix)
	}
	if _, err := r.Discard(len(magic)); err != nil {
		return err
	}

	// Decode as JSON.
	return httputil.DecodeJSON(r, v)
}
