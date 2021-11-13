package auth

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/auth/user"
	"github.com/metagram-net/firehose/db"
)

func Register(r *mux.Router, db *sql.DB) {
	r.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := api.Context()
		defer cancel()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			api.Respond(w, nil, err)
			return
		}
		// Commit the transaction to avoid leaking its memory. It doesn't
		// really matter whether it closes, though, so we can ignore the error
		// safely.
		//nolint:errcheck
		defer tx.Commit()

		u, err := Whoami(ctx, tx, r)
		api.Respond(w, u, err)
	})
}

type User struct {
	ID uuid.UUID `json:"id"`
}

func Whoami(ctx context.Context, tx db.Queryable, r *http.Request) (*User, error) {
	u, err := user.FromRequest(ctx, tx, r)
	if err != nil {
		return nil, err
	}
	return &User{ID: u.ID}, nil
}
