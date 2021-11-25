package auth

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/auth/apikey"
	"github.com/metagram-net/firehose/auth/user"
	"github.com/metagram-net/firehose/db"
	"go.uber.org/zap"
)

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

type User struct {
	ID uuid.UUID `json:"id"`
}

func Whoami(ctx context.Context, log *zap.Logger, tx db.Queryable, r *http.Request) (*User, error) {
	u, err := FromRequest(ctx, log, tx, r)
	if err != nil {
		return nil, err
	}
	return &User{ID: u.ID}, nil
}

// TODO: Replace API key "passwords" with PK registration and request signing.

func FromRequest(ctx context.Context, log *zap.Logger, tx db.Queryable, req *http.Request) (*User, error) {
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
	u, err := user.Find(ctx, tx, key.UserID)
	if err != nil {
		log.Warn("Could not find user by ID", zap.Error(err))
		return nil, ErrInvalidAuthz
	}
	return &User{ID: u.ID}, nil
}
