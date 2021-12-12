package droptags

import (
	"context"

	"github.com/blockloop/scan"
	"github.com/gofrs/uuid"
	"github.com/metagram-net/firehose/db"
	"github.com/metagram-net/firehose/drop/internal/drops"
	"github.com/metagram-net/firehose/drop/internal/tags"
)

type Record struct {
	ID     uuid.UUID `db:"id"`
	DropID uuid.UUID `db:"drop_id"`
	TagID  uuid.UUID `db:"tag_id"`
}

func Insert(ctx context.Context, tx db.Queryable, dr drops.Record, ts []tags.Record) ([]Record, error) {
	q := db.Pq.
		Insert("drop_tags").
		Columns("drop_id", "tag_id").
		Suffix("RETURNING *")
	for _, t := range ts {
		q = q.Values(dr.ID, t.ID)
	}
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	var r []Record
	return r, scan.RowsStrict(&r, rows)
}
