package drop

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/auth"
)

type Handler struct{}

func (Handler) Random(a api.Context, u auth.User, w http.ResponseWriter, _ *http.Request) {
	ctx, log, tx := a.Ctx, a.Log, a.Tx
	d, err := Random(ctx, tx, u)
	api.Respond(log, w, d, err)
}

func (Handler) Next(a api.Context, u auth.User, w http.ResponseWriter, _ *http.Request) {
	ctx, log, tx := a.Ctx, a.Log, a.Tx
	d, err := Next(ctx, tx, u)
	api.Respond(log, w, d, err)
}

func (Handler) Get(a api.Context, u auth.User, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx := a.Ctx, a.Log, a.Tx

	vars := mux.Vars(r)
	id, err := uuid.FromString(vars["id"])
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := Get(ctx, tx, u, id)
	api.Respond(log, w, d, err)
}

func (Handler) List(a api.Context, u auth.User, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx := a.Ctx, a.Log, a.Tx

	var req struct {
		Status Status  `json:"status"`
		Limit  *uint64 `json:"limit"`
	}
	if err := unmarshal(r, &req); err != nil {
		api.Respond(log, w, nil, err)
		return
	}
	// The point of Firehose is to force me to consume or delete the oldest
	// content first. Listing is useful if some content is not currently
	// consumable (videos and PDFs are the usual examples) so you have to
	// "scroll past" a few drops. But the point is to avoid scrolling for a
	// long time to find something, so don't allow large limits here.
	limit := uint64(20)
	if req.Limit != nil && *req.Limit < limit {
		limit = *req.Limit
	}
	ds, err := List(ctx, tx, u, req.Status, limit)

	type res struct {
		Drops []Drop `json:"drops"`
	}
	api.Respond(log, w, res{Drops: ds}, err)
}

func (Handler) Create(a api.Context, u auth.User, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx, clock := a.Ctx, a.Log, a.Tx, a.Clock

	var req struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if err := unmarshal(r, &req); err != nil {
		api.Respond(log, w, nil, err)
		return
	}
	// TODO: Remove this parsing and fall back to basic strings.
	urlp, err := url.Parse(req.URL)
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := Create(ctx, tx, u, req.Title, *urlp, clock.Now())
	api.Respond(log, w, d, err)
}

func (Handler) Update(a api.Context, u auth.User, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx, clock := a.Ctx, a.Log, a.Tx, a.Clock

	vars := mux.Vars(r)
	id, err := uuid.FromString(vars["id"])
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	var req UpdateRequest
	if err := unmarshal(r, &req); err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := Update(ctx, tx, u, id, req, clock.Now())
	api.Respond(log, w, d, err)
}

func (Handler) Delete(a api.Context, u auth.User, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx := a.Ctx, a.Log, a.Tx

	vars := mux.Vars(r)
	id, err := uuid.FromString(vars["id"])
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := Delete(ctx, tx, u, id)
	api.Respond(log, w, d, err)
}

func unmarshal(r *http.Request, v interface{}) error {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
