package user

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/blockloop/scan"
	"github.com/gofrs/uuid"
	"github.com/metagram-net/firehose/db"
)

type Record struct {
	ID    uuid.UUID `db:"id"`
	Email string    `db:"email_address"`
	db.Timestamps
}

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
