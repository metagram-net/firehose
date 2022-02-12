package db

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/blockloop/scan"
	"github.com/gofrs/uuid"
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
	Title *string
	URL   *string
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
	if title := f.Set.Title; title != nil {
		qq = qq.Set("title", title)
	}
	if url := f.Set.URL; url != nil {
		qq = qq.Set("url", url)
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
