package db

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/blockloop/scan"
	"github.com/gofrs/uuid"

	"github.com/metagram-net/firehose/null"
)

var Pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

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

	// TODO: Would CopyFrom work here?
	// https://docs.sqlc.dev/en/latest/howto/insert.html#using-copyfrom

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

type DropUpdateFields struct {
	Select DropUpdateSelect
	Set    DropUpdateSet
}

type DropUpdateSelect struct {
	ID     uuid.UUID `db:"id"`
	UserID uuid.UUID `db:"user_id"`
}

type DropUpdateSet struct {
	Title null.String `db:"title"`
	URL   null.String `db:"url"`
}

type dropSelect struct {
	ID        uuid.UUID      `db:"id"`
	UserID    uuid.UUID      `db:"user_id"`
	Title     sql.NullString `db:"title"`
	URL       string         `db:"url"`
	Status    DropStatus     `db:"status"`
	MovedAt   time.Time      `db:"moved_at"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

func DropUpdate(ctx context.Context, tx DBTX, f DropUpdateFields) (Drop, error) {
	qq := Pq.
		Update("drops").
		Where(sq.Eq{
			"id":      f.Select.ID,
			"user_id": f.Select.UserID,
		}).
		Suffix("RETURNING *")

	// TODO: reflection?
	if title := f.Set.Title; title.Present {
		qq = qq.Set("title", title.Value)
	}
	if url := f.Set.URL; url.Present {
		qq = qq.Set("url", url.Value)
	}

	query, args, err := qq.ToSql()
	if err != nil {
		return Drop{}, err
	}

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return Drop{}, err
	}

	var d dropSelect
	return Drop(d), scan.RowStrict(&d, rows)
}
