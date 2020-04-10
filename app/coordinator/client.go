package coordinator

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/httputil"
)

type Client struct {
	client *http.Client
	url    string
	worker string
}

func NewClient(c *http.Client, url, worker string) *Client {
	return &Client{
		client: c,
		url:    url,
		worker: worker,
	}
}

func (c *Client) Jobs(ctx context.Context) (*JobsResponse, error) {
	// Build request.
	u := c.url + "/workers/" + c.worker + "/jobs"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, nil)
	if err != nil {
		return nil, err
	}

	// Execute the request.
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := httputil.ExpectStatus(res.StatusCode, http.StatusOK, http.StatusCreated); err != nil {
		return nil, err
	}

	// Decode JSON response.
	payload := &JobsResponse{}
	if err := httputil.DecodeJSON(res.Body, payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (c *Client) Start(ctx context.Context, id uuid.UUID) error {
	// Build request.
	u := c.url + "/workers/" + c.worker + "/jobs/" + id.String() + "/start"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, nil)
	if err != nil {
		return err
	}

	// Execute the request.
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if err := httputil.ExpectStatus(res.StatusCode, http.StatusNoContent); err != nil {
		return err
	}

	return nil
}
