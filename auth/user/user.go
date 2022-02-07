package user

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/metagram-net/firehose/db"
)

type User struct {
	ID string `json:"id"`
}

func Create(ctx context.Context, q db.Querier, email string) (*db.User, error) {
	u, err := q.UserCreate(ctx, email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func Find(ctx context.Context, q db.Querier, id uuid.UUID) (*db.User, error) {
	u, err := q.UserFind(ctx, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
