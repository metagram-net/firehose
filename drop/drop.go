package drop

import (
	"context"
	"database/sql"
	"time"

	"github.com/blockloop/scan"
	"github.com/gofrs/uuid"
	"github.com/metagram-net/firehose/article"
	"github.com/metagram-net/firehose/db"
)

type Drop struct {
	ID      string     `json:"id"`
	Title   string     `json:"title"`
	Status  Status     `json:"status"`
	MovedAt *time.Time `json:"movedAt"`
	URL     string     `json:"url"`
}

type Record struct {
	ID        uuid.UUID      `db:"id"`
	Title     sql.NullString `db:"title"`
	Status    Status         `db:"status"`
	MovedAt   sql.NullTime   `db:"moved_at"` // TODO: make non-nullable
	ArticleID uuid.UUID      `db:"article_id"`
	UserID    uuid.UUID      `db:"user_id"`

	db.Timestamps
}

func Random(ctx context.Context, tx db.Queryable) (*Drop, error) {
	var record struct {
		Drop    Record         `db:"drops"`
		Article article.Record `db:"articles"`
	}
	rows, err := tx.QueryContext(ctx, qRandomDrop)
	if err != nil {
		return nil, err
	}
	if err := scan.RowStrict(&record, rows); err != nil {
		return nil, err
	}

	title := firstString(
		record.Drop.Title.String,
		record.Article.Title.String,
		record.Article.URL.String,
	)

	var movedAt *time.Time
	if record.Drop.MovedAt.Valid {
		movedAt = &record.Drop.MovedAt.Time
	}

	return &Drop{
		ID:      record.Drop.ID.String(),
		Title:   title,
		URL:     record.Article.URL.String,
		Status:  record.Drop.Status,
		MovedAt: movedAt,
	}, nil
}

func firstString(strs ...string) string {
	for _, s := range strs {
		if s != "" {
			return s
		}
	}
	return ""
}
