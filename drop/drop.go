package drop

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/db"
	"github.com/metagram-net/firehose/null"
)

type Drop struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	URL     string    `json:"url"`
	Status  Status    `json:"status"`
	MovedAt time.Time `json:"moved_at"`
	Tags    []Tag     `json:"tags"`
}

type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func model(d db.Drop, ts []db.Tag) Drop {
	tags := make([]Tag, 0)
	for _, t := range ts {
		tags = append(tags, Tag{
			ID:   t.ID.String(),
			Name: t.Name,
		})
	}
	return Drop{
		ID:      d.ID.String(),
		Title:   d.Title.String,
		URL:     d.URL,
		Status:  StatusModel(d.Status),
		MovedAt: d.MovedAt,
		Tags:    tags,
	}
}

func Create(ctx context.Context, q db.Queryable, user api.User, title string, url string, tagIDs []uuid.UUID, now time.Time) (Drop, error) {
	var ts []db.Tag
	if len(tagIDs) > 0 {
		var err error
		ts, err = q.TagFindAll(ctx, db.TagFindAllParams{
			UserID: user.ID,
			Ids:    tagIDs,
		})
		if err != nil {
			return Drop{}, err
		}
	}

	d, err := q.DropCreate(ctx, db.DropCreateParams{
		UserID:  user.ID,
		Title:   db.NullString(&title),
		URL:     url,
		Status:  db.DropStatusUnread,
		MovedAt: now,
	})
	if err != nil {
		return Drop{}, err
	}

	_, err = db.DropTagsApply(ctx, q, d, ts)
	if err != nil {
		return Drop{}, err
	}
	return model(d, ts), err
}

type UpdateFields struct {
	Title null.String `db:"title"`
	URL   null.String `db:"url"`
}

func Update(ctx context.Context, q db.Queryable, user api.User, id uuid.UUID, f UpdateFields) (Drop, error) {
	// TODO(tags): Update drop_tags
	d, err := db.DropUpdate(ctx, q, db.DropUpdateFields{
		Select: db.DropUpdateSelect{
			ID:     id,
			UserID: user.ID,
		},
		Set: db.DropUpdateSet(f),
	})
	if err != nil {
		return Drop{}, err
	}
	return model(d, nil), nil
}

func Move(ctx context.Context, q db.Queryable, user api.User, id uuid.UUID, status Status, now time.Time) (Drop, error) {
	d, err := q.DropMove(ctx, db.DropMoveParams{
		UserID:  user.ID,
		ID:      id,
		Status:  status.Model(),
		MovedAt: now,
	})
	if err != nil {
		return Drop{}, err
	}
	return model(d, nil), nil
}

func Delete(ctx context.Context, q db.Queryable, user api.User, id uuid.UUID) (Drop, error) {
	// TODO(tags): Delete drop_tags
	d, err := q.DropDelete(ctx, db.DropDeleteParams{
		UserID: user.ID,
		ID:     id,
	})
	if err != nil {
		return Drop{}, err
	}
	return model(d, nil), nil
}

func Get(ctx context.Context, q db.Queryable, user api.User, id uuid.UUID) (Drop, error) {
	d, err := q.DropFind(ctx, db.DropFindParams{
		UserID: user.ID,
		ID:     id,
	})
	if err != nil {
		return Drop{}, err
	}
	ts, err := q.TagsDrop(ctx, db.TagsDropParams{
		UserID: user.ID,
		ID:     d.ID,
	})
	return model(d, ts), err
}

func Next(ctx context.Context, q db.Queryable, user api.User) (Drop, error) {
	d, err := q.DropNext(ctx, user.ID)
	if err != nil {
		return Drop{}, err
	}
	ts, err := q.TagsDrop(ctx, db.TagsDropParams{
		UserID: user.ID,
		ID:     d.ID,
	})
	return model(d, ts), err
}

func List(ctx context.Context, q db.Queryable, user api.User, s Status, limit int32) ([]Drop, error) {
	ds, err := q.DropList(ctx, db.DropListParams{
		UserID:   user.ID,
		Statuses: []db.DropStatus{s.Model()},
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}
	var dropIDs []uuid.UUID

	tagRows, err := q.TagsDrops(ctx, db.TagsDropsParams{
		UserID:  user.ID,
		DropIds: dropIDs,
	})
	if err != nil {
		return nil, err
	}
	tags := make(map[uuid.UUID][]db.Tag)
	for _, r := range tagRows {
		tags[r.DropID] = append(tags[r.DropID], db.Tag{
			ID:        r.ID,
			UserID:    r.UserID,
			Name:      r.Name,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		})
	}

	// The list of drops should never be nil/null, so always make the slice.
	res := make([]Drop, 0)
	for _, d := range ds {
		res = append(res, model(d, tags[d.ID]))
	}
	return res, err
}
