package coordinator

import (
	"context"
	"io"
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
	payload := &JobsResponse{}
	if err := c.request(ctx, params{
		Method:         http.MethodPost,
		Path:           "/workers/" + c.worker + "/jobs",
		AcceptStatuses: []int{http.StatusOK, http.StatusCreated},
		Payload:        payload,
	}); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *Client) Start(ctx context.Context, id uuid.UUID) error {
	return c.request(ctx, params{
		Method:         http.MethodPut,
		Path:           "/workers/" + c.worker + "/jobs/" + id.String() + "/start",
		AcceptStatuses: []int{http.StatusNoContent},
	})
}

func (c *Client) UploadResult(ctx context.Context, id uuid.UUID, r io.Reader) error {
	return c.request(ctx, params{
		Method:         http.MethodPut,
		Path:           "/workers/" + c.worker + "/jobs/" + id.String() + "/result",
		Body:           r,
		AcceptStatuses: []int{http.StatusNoContent},
	})
}

func (c *Client) Fail(ctx context.Context, id uuid.UUID) error {
	return c.request(ctx, params{
		Method:         http.MethodPut,
		Path:           "/workers/" + c.worker + "/jobs/" + id.String() + "/fail",
		AcceptStatuses: []int{http.StatusNoContent},
	})
}

func (c *Client) Halt(ctx context.Context, id uuid.UUID) error {
	return c.request(ctx, params{
		Method:         http.MethodPut,
		Path:           "/workers/" + c.worker + "/jobs/" + id.String() + "/halt",
		AcceptStatuses: []int{http.StatusNoContent},
	})
}

type params struct {
	Method         string
	Path           string
	Body           io.Reader
	AcceptStatuses []int
	Payload        interface{}
}

func (c *Client) request(ctx context.Context, p params) error {
	// Build request.
	url := c.url + p.Path
	req, err := http.NewRequestWithContext(ctx, p.Method, url, p.Body)
	if err != nil {
		return err
	}

	// Execute the request.
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if err := httputil.ExpectStatus(res.StatusCode, p.AcceptStatuses...); err != nil {
		return err
	}

	// Decode JSON response.
	if p.Payload != nil {
		if err := httputil.DecodeJSON(res.Body, p.Payload); err != nil {
			return err
		}
	}

	return nil
}
