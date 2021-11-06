package article

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Record struct {
	ID        uuid.UUID      `db:"id"`
	Title     sql.NullString `db:"title"`
	URL       sql.NullString `db:"url"` // TODO: make non-nullable
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}
