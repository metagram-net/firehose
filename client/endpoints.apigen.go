// Code generated by apigen; DO NOT EDIT.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/auth"
	"github.com/metagram-net/firehose/drop"
	"github.com/metagram-net/firehose/wellknown"
)

type Fetcher interface {
	Get(ctx context.Context, path string) (*http.Response, error)
	Post(ctx context.Context, path string, body io.Reader) (*http.Response, error)
}

type Endpoints struct {
	WellKnown WellKnown
	Auth      Auth
	Drops     Drops
}

func NewEndpoints(f Fetcher) Endpoints {
	return Endpoints{
		WellKnown: WellKnown{f},
		Auth:      Auth{f},
		Drops:     Drops{f},
	}
}

type WellKnown struct {
	f Fetcher
}

func (g WellKnown) HealthCheck(ctx context.Context) (wellknown.HealthCheckResponse, error) {
	var val wellknown.HealthCheckResponse

	path := "/.well-known/health-check"

	res, err := g.f.Get(ctx, path)
	if err != nil {
		return val, err
	}

	return val, parse(res, &val)
}

type Auth struct {
	f Fetcher
}

func (g Auth) Whoami(ctx context.Context) (auth.User, error) {
	var val auth.User

	path := "/auth/whoami"

	res, err := g.f.Get(ctx, path)
	if err != nil {
		return val, err
	}

	return val, parse(res, &val)
}

type Drops struct {
	f Fetcher
}

func (g Drops) Next(ctx context.Context) (drop.Drop, error) {
	var val drop.Drop

	path := "/v1/drops/next"

	res, err := g.f.Get(ctx, path)
	if err != nil {
		return val, err
	}

	return val, parse(res, &val)
}

func (g Drops) Get(ctx context.Context, params drop.GetParams) (drop.Drop, error) {
	var val drop.Drop

	url, err := (&mux.Route{}).Path("/v1/drops/get/{id}").URL(api.Pairs(params)...)
	if err != nil {
		return val, err
	}
	path := url.String()

	res, err := g.f.Get(ctx, path)
	if err != nil {
		return val, err
	}

	return val, parse(res, &val)
}

func (g Drops) List(ctx context.Context, body drop.ListBody) (drop.ListResponse, error) {
	var val drop.ListResponse

	path := "/v1/drops/list"

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(body); err != nil {
		return val, err
	}
	res, err := g.f.Post(ctx, path, &b)
	if err != nil {
		return val, err
	}

	return val, parse(res, &val)
}

func (g Drops) Create(ctx context.Context, body drop.CreateBody) (drop.Drop, error) {
	var val drop.Drop

	path := "/v1/drops/create"

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(body); err != nil {
		return val, err
	}
	res, err := g.f.Post(ctx, path, &b)
	if err != nil {
		return val, err
	}

	return val, parse(res, &val)
}

func (g Drops) Update(ctx context.Context, body drop.UpdateBody) (drop.Drop, error) {
	var val drop.Drop

	path := "/v1/drops/update"

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(body); err != nil {
		return val, err
	}
	res, err := g.f.Post(ctx, path, &b)
	if err != nil {
		return val, err
	}

	return val, parse(res, &val)
}

func (g Drops) Move(ctx context.Context, body drop.MoveBody) (drop.Drop, error) {
	var val drop.Drop

	path := "/v1/drops/move"

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(body); err != nil {
		return val, err
	}
	res, err := g.f.Post(ctx, path, &b)
	if err != nil {
		return val, err
	}

	return val, parse(res, &val)
}

func (g Drops) Delete(ctx context.Context, body drop.DeleteBody) (drop.Drop, error) {
	var val drop.Drop

	path := "/v1/drops/delete"

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(body); err != nil {
		return val, err
	}
	res, err := g.f.Post(ctx, path, &b)
	if err != nil {
		return val, err
	}

	return val, parse(res, &val)
}