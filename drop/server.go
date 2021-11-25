package drop

import (
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
	"github.com/metagram-net/firehose/clock"
	"go.uber.org/zap"
)

type Server struct {
	log *zap.Logger
	db  *sql.DB
}

func NewServer(log *zap.Logger, db *sql.DB) *Server {
	return &Server{log, db}
}

func (s *Server) Random(w http.ResponseWriter, r *http.Request) { s.authed(w, r, s.random) }
func (s *Server) Next(w http.ResponseWriter, r *http.Request)   { s.authed(w, r, s.next) }
func (s *Server) Get(w http.ResponseWriter, r *http.Request)    { s.authed(w, r, s.get) }
func (s *Server) Create(w http.ResponseWriter, r *http.Request) { s.authed(w, r, s.create) }
func (s *Server) Update(w http.ResponseWriter, r *http.Request) { s.authed(w, r, s.update) }
func (s *Server) Delete(w http.ResponseWriter, r *http.Request) { s.authed(w, r, s.delete) }

type authedHandler func(api.Context, user.Record, http.ResponseWriter, *http.Request)

func (s *Server) authed(w http.ResponseWriter, r *http.Request, next authedHandler) {
	ctx := r.Context()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		api.Respond(s.log, w, nil, err)
		return
	}
	defer func() {
		if err := tx.Commit(); err != nil {
			s.log.Error("Could not commit transaction", zap.Error(err))
		}
	}()

	u, err := user.FromRequest(ctx, s.log, tx, r)
	if err != nil {
		api.Respond(s.log, w, nil, err)
		return
	}

	a := api.Context{
		Ctx:   ctx,
		Log:   s.log,
		Tx:    tx,
		Clock: clock.Freeze(time.Now()),
	}
	next(a, *u, w, r)
}

func unmarshal(r *http.Request, v interface{}) error {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func (s *Server) random(a api.Context, u user.Record, w http.ResponseWriter, _ *http.Request) {
	ctx, log, tx := a.Ctx, a.Log, a.Tx
	d, err := Random(ctx, tx, u)
	api.Respond(log, w, d, err)
}

func (s *Server) next(a api.Context, u user.Record, w http.ResponseWriter, _ *http.Request) {
	ctx, log, tx := a.Ctx, a.Log, a.Tx
	d, err := Next(ctx, tx, u)
	api.Respond(log, w, d, err)
}

func (s *Server) get(a api.Context, u user.Record, w http.ResponseWriter, r *http.Request) {
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

func (s *Server) create(a api.Context, u user.Record, w http.ResponseWriter, r *http.Request) {
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

func (s *Server) update(a api.Context, u user.Record, w http.ResponseWriter, r *http.Request) {
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

func (s *Server) delete(a api.Context, u user.Record, w http.ResponseWriter, r *http.Request) {
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
