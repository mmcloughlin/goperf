package gitiles

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

func (c *Client) Log(ctx context.Context, repo string) (*LogResponse, error) {
	// Build request.
	u := c.base + "/" + repo + "/+log?format=JSON"
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	// Execute the request.
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Parse response body.
	payload := &LogResponse{}
	if err := decodejson(res.Body, payload); err != nil {
		return nil, err
	}

	return payload, nil
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
	r.Discard(len(magic))

	// Decode as JSON.
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()

	if err := d.Decode(v); err != nil {
		return err
	}

	// Should not have trailing data.
	if d.More() {
		return errors.New("unexpected extra data after JSON")
	}

	return nil
}
