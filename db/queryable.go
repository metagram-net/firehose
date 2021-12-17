package db

import (
	"context"
	"database/sql"
)

type Queryable interface {
	Querier
	DBTX
}

func (q *Queries) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return q.db.ExecContext(ctx, query, args...)
}

func (q *Queries) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return q.db.PrepareContext(ctx, query)
}

func (q *Queries) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return q.db.QueryContext(ctx, query, args...)
}

func (q *Queries) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return q.db.QueryRowContext(ctx, query, args...)
}
