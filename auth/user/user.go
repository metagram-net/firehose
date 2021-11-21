package user

import (
	"context"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/blockloop/scan"
	"github.com/gofrs/uuid"
	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/auth/apikey"
	"github.com/metagram-net/firehose/db"
	"go.uber.org/zap"
)

type Record struct {
	ID    uuid.UUID `db:"id"`
	Email string    `db:"email_address"`
	db.Timestamps
}

var (
	ErrMissingAuthz = api.NewError(
		http.StatusUnauthorized,
		"missing_authorization",
		"The Authorization header was missing or the wrong format.",
	)
	ErrInvalidAuthz = api.NewError(
		http.StatusUnauthorized,
		"invalid_authorization",
		"The provided credentials were not valid.",
	)
)

func Create(ctx context.Context, tx db.Queryable, email string) (*Record, error) {
	query, args, err := db.Pq.
		Insert("users").
		SetMap(db.Row{"email_address": email}).
		Suffix("returning *").
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

func FromRequest(ctx context.Context, log *zap.Logger, tx db.Queryable, req *http.Request) (*Record, error) {
	// If the header is missing completely, return a more helpful error.
	if req.Header.Get("Authorization") == "" {
		return nil, ErrMissingAuthz
	}

	// Parse out the user ID and token
	username, password, ok := req.BasicAuth()
	if !ok {
		log.Warn("Invalid basic auth header")
		return nil, ErrInvalidAuthz
	}
	userID, err := uuid.FromString(username)
	if err != nil {
		log.Warn("Could not parse user ID", zap.Error(err))
		return nil, ErrInvalidAuthz
	}
	token, err := apikey.NewPlaintext(password)
	if err != nil {
		log.Warn("Could not parse token", zap.Error(err))
		return nil, ErrInvalidAuthz
	}

	key, err := apikey.Find(ctx, tx, userID, token)
	if err != nil {
		log.Warn("Could not find API key by user ID and token", zap.Error(err))
		return nil, ErrInvalidAuthz
	}
	u, err := Find(ctx, tx, key.UserID)
	if err != nil {
		log.Warn("Could not find user by ID", zap.Error(err))
		return nil, ErrInvalidAuthz
	}
	return u, nil
}
