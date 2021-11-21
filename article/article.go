package article

import (
	"context"
	"database/sql"
	"net/url"
	"time"

	"github.com/blockloop/scan"
	"github.com/google/uuid"
	"github.com/metagram-net/firehose/db"
)

type Record struct {
	ID        uuid.UUID      `db:"id"`
	Title     sql.NullString `db:"title"`
	URL       sql.NullString `db:"url"` // TODO: make non-nullable
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

func Create(ctx context.Context, tx db.Queryable, title string, url url.URL) (*Record, error) {
	query, args, err := db.Pq.
		Insert("articles").
		SetMap(db.Row{
			"title": title,
			"url":   url.String(),
		}).
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
