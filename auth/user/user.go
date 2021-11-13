package user

import (
	"context"
	"log"
	"net/http"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/blockloop/scan"
	"github.com/gofrs/uuid"
	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/auth/apikey"
	"github.com/metagram-net/firehose/db"
)

type Record struct {
	ID uuid.UUID `db:"id"`
	db.Timestamps
}

var (
	ErrMissingAuthz = api.NewError(http.StatusUnauthorized, "missing_authorization", "Invalid Authorization header")
	ErrInvalidAuthz = api.NewError(http.StatusUnauthorized, "invalid_authorization", "Invalid credentials")
)

func Create(ctx context.Context, tx db.Queryable) (*Record, error) {
	query, args, err := db.Pq.Insert("users").Suffix("returning *").ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	var r Record
	return &r, scan.RowStrict(&r, rows)
}

func Find(ctx context.Context, tx db.Queryable, id uuid.UUID) (*Record, error) {
	query, args, err := db.Pq.
		Select("*").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	var r Record
	return &r, scan.RowStrict(&r, rows)
}

func FromRequest(ctx context.Context, tx db.Queryable, req *http.Request) (*Record, error) {
	// Based on net/http.Request.BasicAuth. Changed for Bearer auth
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return nil, ErrMissingAuthz
	}
	prefix := "Bearer "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return nil, ErrMissingAuthz
	}
	token, err := apikey.NewPlaintext(auth[len(prefix):])
	if err != nil {
		// TODO(start-here): real logging
		log.Print("Could not parse token")
		return nil, ErrInvalidAuthz
	}
	key, err := apikey.Find(ctx, tx, token)
	if err != nil {
		return nil, err
	}
	return Find(ctx, tx, key.UserID)
}
