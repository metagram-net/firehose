package drop

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/auth/user"
	"github.com/metagram-net/firehose/db"
	"go.uber.org/zap"
)

type server struct {
	log *zap.Logger
	db  *sql.DB
}

func (s *server) Random(hctx api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx := hctx.Ctx, hctx.Log, hctx.Tx

	u, err := user.FromRequest(ctx, log, tx, r)
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := Random(ctx, tx, *u)
	api.Respond(log, w, d, err)
}

func (s *server) Next(hctx api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx := hctx.Ctx, hctx.Log, hctx.Tx

	u, err := user.FromRequest(ctx, log, tx, r)
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := Next(ctx, tx, *u)
	api.Respond(log, w, d, err)
}

func (s *server) Get(hctx api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx := hctx.Ctx, hctx.Log, hctx.Tx

	u, err := user.FromRequest(ctx, log, tx, r)
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.FromString(vars["id"])
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := Get(ctx, tx, *u, id)
	api.Respond(log, w, d, err)
}

func (s *server) Create(hctx api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx, clock := hctx.Ctx, hctx.Log, hctx.Tx, hctx.Clock

	u, err := user.FromRequest(ctx, log, tx, r)
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	var req struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}
	if err := json.Unmarshal(b, &req); err != nil {
		api.Respond(log, w, nil, err)
		return
	}
	// TODO: Remove this parsing and fall back to basic strings.
	urlp, err := url.Parse(req.URL)
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := Create(ctx, tx, *u, req.Title, *urlp, clock.Now())
	api.Respond(log, w, d, err)
}

func (s *server) Update(hctx api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx, clock := hctx.Ctx, hctx.Log, hctx.Tx, hctx.Clock

	u, err := user.FromRequest(ctx, log, tx, r)
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.FromString(vars["id"])
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	var req UpdateRequest
	b, err := io.ReadAll(r.Body)
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}
	if err := json.Unmarshal(b, &req); err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := Update(ctx, tx, *u, id, req, clock.Now())
	api.Respond(log, w, d, err)
}

func (s *server) Delete(hctx api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	ctx, log, tx := hctx.Ctx, hctx.Log, hctx.Tx

	u, err := user.FromRequest(ctx, log, tx, r)
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.FromString(vars["id"])
	if err != nil {
		api.Respond(log, w, nil, err)
		return
	}

	d, err := Delete(ctx, tx, *u, id)
	api.Respond(log, w, d, err)
}

func Register(r *mux.Router, db *sql.DB, log *zap.Logger) {
	// TODO: If s already knows the log and db, why use api.Handle?
	s := server{log, db}

	r.Methods(http.MethodGet).Path("/random").Handler(api.Handle(db, log, s.Random))
	r.Methods(http.MethodGet).Path("/next").Handler(api.Handle(db, log, s.Next))
	r.Methods(http.MethodGet).Path("/get/{id}").Handler(api.Handle(db, log, s.Get))

	r.Methods(http.MethodPost).Path("/create").Handler(api.Handle(db, log, s.Create))
	r.Methods(http.MethodPost).Path("/update/{id}").Handler(api.Handle(db, log, s.Update))
	r.Methods(http.MethodPost).Path("/delete/{id}").Handler(api.Handle(db, log, s.Delete))
}

func Random(ctx context.Context, tx db.Queryable, user user.Record) (Drop, error) {
	d, err := ForUser(user.ID).Random(ctx, tx)
	if err != nil {
		return Drop{}, err
	}
	return d.Model(), err
}

func Create(ctx context.Context, tx db.Queryable, user user.Record, title string, url url.URL, now time.Time) (Drop, error) {
	d, err := ForUser(user.ID).Create(ctx, tx, title, url, now)
	if err != nil {
		return Drop{}, err
	}
	return d.Model(), err
}

type UpdateRequest struct {
	Title  *string `json:"title"`
	URL    *string `json:"url"`
	Status *Status `json:"status"`
}

func Update(ctx context.Context, tx db.Queryable, user user.Record, id uuid.UUID, req UpdateRequest, now time.Time) (Drop, error) {
	f := Fields{
		Title:  req.Title,
		URL:    req.URL,
		Status: req.Status,
	}
	// Mark when the status changed so streams act more like FIFO queues.
	if f.Status != nil {
		f.MovedAt = &now
	}
	d, err := ForUser(user.ID).Update(ctx, tx, id, f)
	if err != nil {
		return Drop{}, err
	}
	return d.Model(), err
}

func Delete(ctx context.Context, tx db.Queryable, user user.Record, id uuid.UUID) (Drop, error) {
	d, err := ForUser(user.ID).Delete(ctx, tx, id)
	if err != nil {
		return Drop{}, err
	}
	return d.Model(), err
}

func Get(ctx context.Context, tx db.Queryable, user user.Record, id uuid.UUID) (Drop, error) {
	d, err := ForUser(user.ID).Find(ctx, tx, id)
	if err != nil {
		return Drop{}, err
	}
	return d.Model(), err
}

func Next(ctx context.Context, tx db.Queryable, user user.Record) (Drop, error) {
	d, err := ForUser(user.ID).Next(ctx, tx)
	if err != nil {
		return Drop{}, err
	}
	return d.Model(), err
}
