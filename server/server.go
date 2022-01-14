package server

import (
	"database/sql"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/drop"
	"github.com/metagram-net/firehose/wellknown"
	"go.uber.org/zap"
)

func New(log *zap.Logger, db *sql.DB) *mux.Router {
	r := mux.NewRouter()

	r.Use(api.NewLogMiddleware(log))

	srv := api.NewServer(log, db)

	r.Methods(http.MethodGet).Path("/.well-known/health-check").HandlerFunc(wellknown.HealthCheck)

	r.Methods(http.MethodGet).Path("/auth/whoami").HandlerFunc(srv.Authed(whoami))

	var drops Drops
	r.Methods(http.MethodGet).Path("/v1/drops/next").HandlerFunc(srv.Authed(drops.next))
	r.Methods(http.MethodGet).Path("/v1/drops/get/{id}").HandlerFunc(srv.Authed(drops.get))
	r.Methods(http.MethodPost).Path("/v1/drops/list").HandlerFunc(srv.Authed(drops.list))
	r.Methods(http.MethodPost).Path("/v1/drops/create").HandlerFunc(srv.Authed(drops.create))
	r.Methods(http.MethodPost).Path("/v1/drops/update/{id}").HandlerFunc(srv.Authed(drops.update))
	r.Methods(http.MethodPost).Path("/v1/drops/move/{id}").HandlerFunc(srv.Authed(drops.move))
	r.Methods(http.MethodPost).Path("/v1/drops/delete/{id}").HandlerFunc(srv.Authed(drops.delete))

	r.NotFoundHandler = notFound(log)
	r.MethodNotAllowedHandler = methodNotAllowed(log)

	return r
}

func whoami(a api.Context, u api.User, w http.ResponseWriter, r *http.Request) {
	api.Respond(a.Log, w, u, nil)
}

func notFound(log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		api.Respond(log, w, nil, api.ErrNotFound)
	}
}

func methodNotAllowed(log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		api.Respond(log, w, nil, api.ErrMethodNotAllowed)
	}
}

type Drops struct{}

func (Drops) next(a api.Context, u api.User, w http.ResponseWriter, _ *http.Request) {
	ctx, log, tx := a.Ctx, a.Log, a.Tx
	d, err := drop.Next(ctx, tx, u)
	api.Respond(log, w, d, err)
}

func (Drops) get(a api.Context, u api.User, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx := a.Ctx, a.Log, a.Tx

	vars := mux.Vars(r)
	id, err := uuid.FromString(vars["id"])
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := drop.Get(ctx, tx, u, id)
	api.Respond(log, w, d, err)
}

type ListRequest struct {
	Status drop.Status `json:"status"`
	Limit  *int32      `json:"limit"`
	// TODO(tags): Implement tags
	// Tags   []uuid.UUID `json:"tags"`
}

func (Drops) list(a api.Context, u api.User, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx := a.Ctx, a.Log, a.Tx

	var req ListRequest
	if err := api.Parse(r, &req); err != nil {
		api.Respond(log, w, nil, err)
		return
	}
	// The point of Firehose is to force me to consume or delete the oldest
	// content first. Listing is useful if some content is not currently
	// consumable (videos and PDFs are the usual examples) so you have to
	// "scroll past" a few drops. But the point is to avoid scrolling for a
	// long time to find something, so don't allow large limits here.
	limit := int32(20)
	if req.Limit != nil && *req.Limit < limit {
		limit = *req.Limit
	}
	ds, err := drop.List(ctx, tx, u, req.Status, limit)

	type res struct {
		Drops []drop.Drop `json:"drops"`
	}
	api.Respond(log, w, res{Drops: ds}, err)
}

type CreateRequest struct {
	Title  string      `json:"title"`
	URL    string      `json:"url"`
	TagIDs []uuid.UUID `json:"tag_ids"`
}

func (Drops) create(a api.Context, u api.User, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx, clock := a.Ctx, a.Log, a.Tx, a.Clock

	var req CreateRequest
	if err := api.Parse(r, &req); err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := drop.Create(ctx, tx, u, req.Title, req.URL, req.TagIDs, clock.Now())
	api.Respond(log, w, d, err)
}

type UpdateRequest struct {
	Title *string `json:"title"`
	URL   string  `json:"url"`
}

func (Drops) update(a api.Context, u api.User, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx := a.Ctx, a.Log, a.Tx

	vars := mux.Vars(r)
	id, err := uuid.FromString(vars["id"])
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	var req UpdateRequest
	if err := api.Parse(r, &req); err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := drop.Update(ctx, tx, u, id, drop.UpdateRequest{
		Title: req.Title,
		URL:   req.URL,
	})
	api.Respond(log, w, d, err)
}

type MoveRequest struct {
	Status drop.Status `json:"status"`
}

func (Drops) move(a api.Context, u api.User, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx, clock := a.Ctx, a.Log, a.Tx, a.Clock

	vars := mux.Vars(r)
	id, err := uuid.FromString(vars["id"])
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	var req MoveRequest
	if err := api.Parse(r, &req); err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := drop.Move(ctx, tx, u, id, req.Status, clock.Now())
	api.Respond(log, w, d, err)
}

func (Drops) delete(a api.Context, u api.User, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx := a.Ctx, a.Log, a.Tx

	vars := mux.Vars(r)
	id, err := uuid.FromString(vars["id"])
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := drop.Delete(ctx, tx, u, id)
	api.Respond(log, w, d, err)
}
