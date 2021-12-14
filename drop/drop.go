package drop

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/db"
	"github.com/metagram-net/firehose/drop/internal/drops"
	"github.com/metagram-net/firehose/drop/internal/droptags"
	"github.com/metagram-net/firehose/drop/internal/tags"
)

type Drop struct {
	ID      string     `json:"id"`
	Title   string     `json:"title"`
	URL     string     `json:"url"`
	Status  Status     `json:"status"`
	MovedAt *time.Time `json:"moved_at"` // TODO: make non-nullable
	Tags    []Tag      `json:"tags"`
}

type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Status = drops.Status

const (
	StatusUnread = drops.StatusUnread
	StatusRead   = drops.StatusRead
	StatusSaved  = drops.StatusSaved
)

func model(dr drops.Record, trs []tags.Record) Drop {
	tags := make([]Tag, 0)
	for _, t := range trs {
		tags = append(tags, Tag{
			ID:   t.ID.String(),
			Name: t.Name,
		})
	}
	return Drop{
		ID:      dr.ID.String(),
		Title:   dr.Title.String,
		URL:     dr.URL,
		Status:  dr.Status,
		MovedAt: nullTime(dr.MovedAt),
		Tags:    tags,
	}
}

func nullTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func Create(ctx context.Context, tx db.Queryable, user api.User, title string, url string, tagIDs []uuid.UUID, now time.Time) (Drop, error) {
	var ts []tags.Record
	if len(tagIDs) > 0 {
		var err error
		ts, err = tags.User(user.ID).FindAll(ctx, tx, tagIDs)
		if err != nil {
			return Drop{}, err
		}
	}

	d, err := drops.User(user.ID).Create(ctx, tx, title, url, now)
	if err != nil {
		return Drop{}, err
	}

	if len(ts) > 0 {
		_, err := droptags.Insert(ctx, tx, *d, ts)
		if err != nil {
			return Drop{}, err
		}
	}
	return model(*d, ts), err
}

type UpdateRequest struct {
	Title  *string `json:"title"`
	URL    *string `json:"url"`
	Status *Status `json:"status"`
}

func Update(ctx context.Context, tx db.Queryable, user api.User, id uuid.UUID, req UpdateRequest, now time.Time) (Drop, error) {
	f := drops.Fields{
		Title:  req.Title,
		URL:    req.URL,
		Status: req.Status,
	}
	// Mark when the status changed so streams act more like FIFO queues.
	if f.Status != nil {
		f.MovedAt = &now
	}
	d, err := drops.User(user.ID).Update(ctx, tx, id, f)
	if err != nil {
		return Drop{}, err
	}
	// TODO(tags): Update tags
	return model(*d, nil), err
}

func Delete(ctx context.Context, tx db.Queryable, user api.User, id uuid.UUID) (Drop, error) {
	// TODO(tags): Delete drop_tags
	d, err := drops.User(user.ID).Delete(ctx, tx, id)
	if err != nil {
		return Drop{}, err
	}
	return model(*d, nil), err
}

func Get(ctx context.Context, tx db.Queryable, user api.User, id uuid.UUID) (Drop, error) {
	d, err := drops.User(user.ID).Find(ctx, tx, id)
	if err != nil {
		return Drop{}, err
	}
	return model(*d, nil), err
}

func Next(ctx context.Context, tx db.Queryable, user api.User) (Drop, error) {
	d, err := drops.User(user.ID).Next(ctx, tx)
	if err != nil {
		return Drop{}, err
	}
	return model(*d, nil), err
}

func List(ctx context.Context, tx db.Queryable, user api.User, s Status, limit uint64) ([]Drop, error) {
	ds, err := drops.User(user.ID).List(ctx, tx, s, limit)
	if err != nil {
		return nil, err
	}
	// The list of drops should never be nil/null, so always make the slice.
	res := make([]Drop, 0)
	for _, d := range ds {
		res = append(res, model(d, nil))
	}
	return res, err
}
