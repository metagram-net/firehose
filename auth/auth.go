package auth

import (
	"context"
	"fmt"

	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/auth/apikey"
	"github.com/metagram-net/firehose/auth/user"
	"github.com/metagram-net/firehose/db"
)

type Registration struct {
	User      db.User
	ApiKey    db.ApiKey
	Plaintext apikey.Plaintext
}

const DefaultKeyName = "Default"

func RegisterUser(ctx context.Context, tx db.DBTX, email string) (*Registration, error) {
	q := db.New(tx)

	user, err := user.Create(ctx, q, email)
	if err != nil {
		return nil, fmt.Errorf("could not create user: %w", err)
	}

	plain, key, err := apikey.Create(ctx, q, DefaultKeyName, user.ID)
	if err != nil {
		return nil, fmt.Errorf("could not create API key: %w", err)
	}

	return &Registration{
		User:      *user,
		ApiKey:    *key,
		Plaintext: *plain,
	}, nil
}

type User struct {
	ID string `json:"id"`
}

type Handler struct{}

//nolint:unparam // The always-nil error is intentional.
func (Handler) Whoami(_ api.Context, u api.User) (User, error) {
	return User{ID: u.ID.String()}, nil
}
