package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
)

var Pq = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// TODO: Use interfaces to model ReadOnly vs. ReadWrite

type Queryable interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type Timestamps struct {
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Row map[string]interface{}
