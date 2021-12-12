package tags

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/blockloop/scan"
	"github.com/gofrs/uuid"
	"github.com/metagram-net/firehose/db"
)

type Record struct {
	ID     uuid.UUID `db:"id"`
	UserID uuid.UUID `db:"user_id"`
	Name   string    `db:"name"`
}

type UserScope struct {
	id uuid.UUID
}

func User(id uuid.UUID) UserScope {
	return UserScope{id: id}
}

func (u UserScope) _select() sq.SelectBuilder {
	return db.Pq.
		Select("*").
		From("tags").
		Where(sq.Eq{"user_id": u.id})
}

func (u UserScope) List(ctx context.Context, tx db.Queryable) ([]Record, error) {
	query, args, err := u.
		_select().
		ToSql()
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

func (u UserScope) FindAll(ctx context.Context, tx db.Queryable, ids []uuid.UUID) ([]Record, error) {
	query, args, err := u.
		_select().
		Where(sq.Eq{"id": ids}).
		ToSql()
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

func (u UserScope) Create(ctx context.Context, tx db.Queryable, name string) (*Record, error) {
	// TODO: struct New that can turn itself into a row by struct tags
	query, args, err := db.Pq.
		Insert("tags").
		SetMap(db.Row{
			"user_id": u.id,
			"name":    name,
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

func (u UserScope) Delete(ctx context.Context, tx db.Queryable, id uuid.UUID) (*Record, error) {

	query, args, err := db.Pq.
		Delete("tags").
		Where(sq.Eq{
			"id":      id,
			"user_id": u.id,
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
