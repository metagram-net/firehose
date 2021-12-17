package db

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/blockloop/scan"
	"github.com/gofrs/uuid"
)

var Pq = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func mustColumns(cols []string, err error) []string {
	if err != nil {
		panic(err)
	}
	return cols
}

type dropTagSelect struct {
	ID        uuid.UUID `db:"id"`
	DropID    uuid.UUID `db:"drop_id"`
	TagID     uuid.UUID `db:"tag_id"`
	CreatedAt time.Time `db:"created_at"`
}

type dropTagInsert struct {
	DropID uuid.UUID `db:"drop_id"`
	TagID  uuid.UUID `db:"tag_id"`
}

var dropTagsInsert []string = mustColumns(scan.ColumnsStrict(new(dropTagInsert)))

func DropTagsApply(ctx context.Context, tx DBTX, d Drop, ts []Tag) ([]DropTag, error) {
	// Building an insert statement with no values causes an error, so return
	// early instead.
	if len(ts) == 0 {
		return nil, nil
	}

	q := Pq.
		Insert("drop_tags").
		Columns(dropTagsInsert...).
		Suffix("RETURNING *")

	for _, t := range ts {
		dt := dropTagInsert{
			DropID: d.ID,
			TagID:  t.ID,
		}
		vs, err := scan.Values(dropTagsInsert, dt)
		if err != nil {
			return nil, err
		}
		q = q.Values(vs...)
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	var rs []dropTagSelect
	if err := scan.RowsStrict(&rs, rows); err != nil {
		return nil, err
	}

	var dts []DropTag
	for _, dt := range rs {
		dts = append(dts, DropTag(dt))
	}
	return dts, nil
}
