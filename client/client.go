package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/metagram-net/firehose/api"
)

type Client struct {
	baseURL url.URL
	userID  string
	apiKey  string

	client *http.Client

	Endpoints
}

func New(opts ...Option) (*Client, error) {
	c := &Client{
		client: &http.Client{},
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	c.Endpoints = NewEndpoints(c)
	return c, nil
}

type Option func(*Client) error

func WithAuth(userID, apiKey string) Option {
	return func(c *Client) error {
		c.userID = userID
		c.apiKey = apiKey
		return nil
	}
}

func WithBaseURL(baseURL string) Option {
	return func(c *Client) error {
		// This base must end in a trailing slash so later Parse calls actually use it.
		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}
		u, err := url.Parse(baseURL)
		c.baseURL = *u
		if err != nil {
			return fmt.Errorf("base URL: %w", err)
		}
		return nil
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		c.client = client
		return nil
	}
}

// parse unmarshals the response body into v. If the response is an HTTP error,
// parse instead tries to unmarshal the body into an api.Error and return it.
// In either case, if unmarshaling fails, this returns the JSON error.
func parse(r *http.Response, v interface{}) error {
	if r.StatusCode >= 400 {
		var e api.Error
		err := json.NewDecoder(r.Body).Decode(&e)
		if err != nil {
			return fmt.Errorf("parse error response: %w", err)
		}
		return e
	}

	return json.NewDecoder(r.Body).Decode(&v)
}

func (c Client) Get(ctx context.Context, path string) (*http.Response, error) {
	url, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.userID, c.apiKey)
	return c.client.Do(req)
}

func (c Client) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	url, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.userID, c.apiKey)
	return c.client.Do(req)
}
