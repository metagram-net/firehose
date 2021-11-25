package drop

import (
	"context"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/metagram-net/firehose/auth/user"
	"github.com/metagram-net/firehose/db"
)

func Random(ctx context.Context, tx db.Queryable, user user.Record) (Drop, error) {
	d, err := ForUser(user.ID).Random(ctx, tx)
	if err != nil {
		return Drop{}, err
	}
	return d.Model(), err
}

func Create(ctx context.Context, tx db.Queryable, user user.Record, title string, url url.URL, now time.Time) (Drop, error) {
	d, err := ForUser(user.ID).Create(ctx, tx, title, url, now)
	if err != nil {
		return Drop{}, err
	}
	return d.Model(), err
}

type UpdateRequest struct {
	Title  *string `json:"title"`
	URL    *string `json:"url"`
	Status *Status `json:"status"`
}

func Update(ctx context.Context, tx db.Queryable, user user.Record, id uuid.UUID, req UpdateRequest, now time.Time) (Drop, error) {
	f := Fields{
		Title:  req.Title,
		URL:    req.URL,
		Status: req.Status,
	}
	// Mark when the status changed so streams act more like FIFO queues.
	if f.Status != nil {
		f.MovedAt = &now
	}
	d, err := ForUser(user.ID).Update(ctx, tx, id, f)
	if err != nil {
		return Drop{}, err
	}
	return d.Model(), err
}

func Delete(ctx context.Context, tx db.Queryable, user user.Record, id uuid.UUID) (Drop, error) {
	d, err := ForUser(user.ID).Delete(ctx, tx, id)
	if err != nil {
		return Drop{}, err
	}
	return d.Model(), err
}

func Get(ctx context.Context, tx db.Queryable, user user.Record, id uuid.UUID) (Drop, error) {
	d, err := ForUser(user.ID).Find(ctx, tx, id)
	if err != nil {
		return Drop{}, err
	}
	return d.Model(), err
}

func Next(ctx context.Context, tx db.Queryable, user user.Record) (Drop, error) {
	d, err := ForUser(user.ID).Next(ctx, tx)
	if err != nil {
		return Drop{}, err
	}
	return d.Model(), err
}
