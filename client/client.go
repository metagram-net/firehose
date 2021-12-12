package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/metagram-net/firehose/drop"
)

type Client struct {
	Drops Drops

	urlBase url.URL
	userID  string
	apiKey  string

	client *http.Client
}

func New(urlBase, userID, apiKey string) (*Client, error) {
	// This base must end in a trailing slash so later Parse calls actually use it.
	if !strings.HasSuffix(urlBase, "/") {
		urlBase += "/"
	}
	u, err := url.Parse(urlBase)
	if err != nil {
		return nil, err
	}
	c := &Client{
		urlBase: *u,
		userID:  userID,
		apiKey:  apiKey,

		client: http.DefaultClient,
	}
	c.Drops = Drops{client: c}
	return c, nil
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
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
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

type Drops struct {
	client *Client
}

func (d Drops) Create(ctx context.Context, title string, url string) (*drop.Drop, error) {
	type request struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	req := request{Title: title, URL: url}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(req); err != nil {
		return nil, err
	}

	res, err := d.client.Post(ctx, "drops/create", &body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var drop drop.Drop
	return &drop, json.NewDecoder(res.Body).Decode(&drop)
}
