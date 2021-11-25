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

func Register(r *mux.Router, db *sql.DB, log *zap.Logger) {
	r.Methods(http.MethodGet).Path("/random").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			api.Respond(log, w, nil, err)
			return
		}
		defer func() {
			if err := tx.Commit(); err != nil {
				log.Error("Could not commit transaction", zap.Error(err))
			}
		}()

		u, err := user.FromRequest(ctx, log, tx, r)
		if err != nil {
			api.Respond(log, w, nil, err)
			return
		}

		d, err := Random(ctx, tx, *u)
		api.Respond(log, w, d, err)
	})

	r.Methods(http.MethodPost).Path("/create").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			api.Respond(log, w, nil, err)
			return
		}
		defer func() {
			if err := tx.Commit(); err != nil {
				log.Error("Could not commit transaction", zap.Error(err))
			}
		}()

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

		d, err := Create(ctx, tx, *u, req.Title, *urlp, time.Now())
		api.Respond(log, w, d, err)
	})

	r.Methods(http.MethodPost).Path("/update/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		vars := mux.Vars(r)
		id, err := uuid.FromString(vars["id"])
		if err != nil {
			api.Respond(log, w, nil, err)
			return
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			api.Respond(log, w, nil, err)
			return
		}
		defer func() {
			if err := tx.Commit(); err != nil {
				log.Error("Could not commit transaction", zap.Error(err))
			}
		}()

		u, err := user.FromRequest(ctx, log, tx, r)
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

		d, err := Update(ctx, tx, *u, id, req, time.Now())
		api.Respond(log, w, d, err)
	})

	r.Methods(http.MethodPost).Path("/delete/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		vars := mux.Vars(r)
		id, err := uuid.FromString(vars["id"])
		if err != nil {
			api.Respond(log, w, nil, err)
			return
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			api.Respond(log, w, nil, err)
			return
		}
		defer func() {
			if err := tx.Commit(); err != nil {
				log.Error("Could not commit transaction", zap.Error(err))
			}
		}()

		u, err := user.FromRequest(ctx, log, tx, r)
		if err != nil {
			api.Respond(log, w, nil, err)
			return
		}

		d, err := Delete(ctx, tx, *u, id)
		api.Respond(log, w, d, err)
	})

	r.Methods(http.MethodGet).Path("/get/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		vars := mux.Vars(r)
		id, err := uuid.FromString(vars["id"])
		if err != nil {
			api.Respond(log, w, nil, err)
			return
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			api.Respond(log, w, nil, err)
			return
		}
		defer func() {
			if err := tx.Commit(); err != nil {
				log.Error("Could not commit transaction", zap.Error(err))
			}
		}()

		u, err := user.FromRequest(ctx, log, tx, r)
		if err != nil {
			api.Respond(log, w, nil, err)
			return
		}

		d, err := Get(ctx, tx, *u, id)
		api.Respond(log, w, d, err)
	})

	r.Methods(http.MethodGet).Path("/next").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			api.Respond(log, w, nil, err)
			return
		}
		defer func() {
			if err := tx.Commit(); err != nil {
				log.Error("Could not commit transaction", zap.Error(err))
			}
		}()

		u, err := user.FromRequest(ctx, log, tx, r)
		if err != nil {
			api.Respond(log, w, nil, err)
			return
		}

		d, err := Next(ctx, tx, *u)
		api.Respond(log, w, d, err)
	})
}

// TODO: I sense a ({ctx, tx, clock}, user, request) pattern forming.

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
