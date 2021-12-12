package drops

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/blockloop/scan"
	"github.com/gofrs/uuid"
	"github.com/metagram-net/firehose/db"
)

type Record struct {
	ID      uuid.UUID      `db:"id"`
	UserID  uuid.UUID      `db:"user_id"`
	Title   sql.NullString `db:"title"` // TODO: make non-nullable
	URL     string         `db:"url"`
	Status  Status         `db:"status"`
	MovedAt sql.NullTime   `db:"moved_at"` // TODO: make non-nullable
}

type Fields struct {
	Title   *string    `db:"title"`
	URL     *string    `db:"url"`
	Status  *Status    `db:"status"`
	MovedAt *time.Time `db:"moved_at"` // TODO: make non-nullable
}

func (f Fields) row() db.Row {
	r := make(db.Row)
	// TODO: This seems like a good place to try reflection and struct tags.
	if f.MovedAt != nil {
		r["moved_at"] = *f.MovedAt
	}
	if f.Status != nil {
		r["status"] = *f.Status
	}
	if f.Title != nil {
		r["title"] = *f.Title
	}
	if f.URL != nil {
		r["url"] = *f.URL
	}
	return r
}

type UserScope struct {
	id uuid.UUID
}

func User(id uuid.UUID) UserScope {
	return UserScope{id: id}
}

// TODO: Filter struct

func (u UserScope) _select() sq.SelectBuilder {
	return db.Pq.
		Select("*").
		From("drops").
		Where(sq.Eq{"user_id": u.id})
}

func (u UserScope) Find(ctx context.Context, tx db.Queryable, id uuid.UUID) (*Record, error) {
	query, args, err := u.
		_select().
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

func (u UserScope) List(ctx context.Context, tx db.Queryable, s Status, limit uint64) ([]Record, error) {
	query, args, err := u.
		_select().
		Where(sq.Eq{"status": s}).
		OrderBy("moved_at ASC").
		Limit(limit).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	var rs []Record
	return rs, scan.RowsStrict(&rs, rows)
}

func (u UserScope) Next(ctx context.Context, tx db.Queryable) (*Record, error) {
	query, args, err := u.
		_select().
		Where(sq.Eq{"drops.status": StatusUnread}).
		OrderBy("drops.moved_at ASC").
		Limit(1).
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

func (u UserScope) Create(ctx context.Context, tx db.Queryable, title string, url string, now time.Time) (*Record, error) {
	// TODO: struct New that can turn itself into a row by struct tags
	query, args, err := db.Pq.
		Insert("drops").
		SetMap(db.Row{
			"moved_at": now,
			"status":   StatusUnread,
			"title":    title,
			"url":      url,
			"user_id":  u.id,
		}).
		Suffix("RETURNING *").
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

func (u UserScope) Update(ctx context.Context, tx db.Queryable, id uuid.UUID, f Fields) (*Record, error) {
	query, args, err := db.Pq.
		Update("drops").
		Where(sq.Eq{
			"id":      id,
			"user_id": u.id,
		}).
		SetMap(f.row()).
		Suffix("RETURNING *").
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

func (u UserScope) Delete(ctx context.Context, tx db.Queryable, id uuid.UUID) (*Record, error) {
	query, args, err := db.Pq.
		Delete("drops").
		Where(sq.Eq{
			"id":      id,
			"user_id": u.id,
		}).
		Suffix("RETURNING *").
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
