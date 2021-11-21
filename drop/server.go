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
	r.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := api.Context()
		defer cancel()

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

	r.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := api.Context()
		defer cancel()

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

	r.HandleFunc("/update/{id}", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := api.Context()
		defer cancel()

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
