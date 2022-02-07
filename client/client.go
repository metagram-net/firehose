package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

// TODO: Turn error statuses into error values

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
