package rest

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// TODO: I'm sure there's real Go REST client out there lol

type Client struct {
	urlBase url.URL
	userID  string
	apiKey  string

	client *http.Client
}

func NewClient(urlBase, userID, apiKey string) (*Client, error) {
	// This base must end in a trailing slash so Parse calls actually use it.
	if !strings.HasSuffix(urlBase, "/") {
		urlBase += "/"
	}
	u, err := url.Parse(urlBase)
	if err != nil {
		return nil, err
	}
	return &Client{
		urlBase: *u,
		userID:  userID,
		apiKey:  apiKey,

		client: http.DefaultClient,
	}, nil
}

// TODO: Turn error statuses into error values

func (c Client) Get(ctx context.Context, path string) (*http.Response, error) {
	url, err := c.urlBase.Parse(path)
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
	url, err := c.urlBase.Parse(path)
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
