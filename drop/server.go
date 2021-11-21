package drop

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

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
		urlp, err := url.Parse(req.URL)
		if err != nil {
			api.Respond(log, w, nil, err)
			return
		}

		d, err := Create(ctx, tx, *u, req.Title, *urlp, time.Now())
		api.Respond(log, w, d, err)
	})
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
