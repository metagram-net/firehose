package auth

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/metagram-net/firehose/auth/user"
	"github.com/metagram-net/firehose/db"
	"go.uber.org/zap"
)

type User struct {
	ID uuid.UUID `json:"id"`
}

func Whoami(ctx context.Context, log *zap.Logger, tx db.Queryable, r *http.Request) (*User, error) {
	u, err := user.FromRequest(ctx, log, tx, r)
	if err != nil {
		return nil, err
	}
	return &User{ID: u.ID}, nil
}
