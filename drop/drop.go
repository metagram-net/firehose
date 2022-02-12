package drop

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/blockloop/scan"
	"github.com/gofrs/uuid"

	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/db"
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
	return loadOne(ctx, q, user, d)
}

type UpdateFields struct {
	Title *string
	URL   *string
	Tags  *[]uuid.UUID
}

func Update(ctx context.Context, q db.Queryable, user api.User, id uuid.UUID, f UpdateFields) (Drop, error) {
	d, err := db.DropUpdate(ctx, q, db.DropUpdateFields{
		Select: db.DropUpdateSelect{
			ID:     id,
			UserID: user.ID,
		},
		Set: db.DropUpdateSet{
			Title: f.Title,
			URL:   f.URL,
		},
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Drop{}, api.NoResourceError("drop", id.String())
	}
	if err != nil {
		return Drop{}, err
	}

	if f.Tags != nil {
		_, err := q.DropTagsIntersect(ctx, db.DropTagsIntersectParams{
			DropID: d.ID,
			TagIds: *f.Tags,
		})
		if err != nil {
			return Drop{}, err
		}

		// TODO: Combine this into one query.
		for _, tagID := range *f.Tags {
			_, err := q.DropTagApply(ctx, db.DropTagApplyParams{
				DropID: d.ID,
				TagID:  tagID,
			})
			if err != nil {
				return Drop{}, err
			}
		}
	}

	return loadOne(ctx, q, user, d)
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
	return loadOne(ctx, q, user, d)
}

func Delete(ctx context.Context, q db.Queryable, user api.User, id uuid.UUID) (Drop, error) {
	// Load the tags we need to render the drop before deleting the drop_tags
	// references.
	tags, err := q.TagsDrop(ctx, db.TagsDropParams{
		UserID: user.ID,
		ID:     id,
	})
	if err != nil {
		return Drop{}, err
	}

	// Delete all the drop tags by intersecting with the empty set. These
	// references need to be removed before the drop can be deleted.
	_, err = q.DropTagsIntersect(ctx, db.DropTagsIntersectParams{
		DropID: id,
		TagIds: []uuid.UUID{},
	})
	if err != nil {
		return Drop{}, err
	}

	d, err := q.DropDelete(ctx, db.DropDeleteParams{
		UserID: user.ID,
		ID:     id,
	})
	if err != nil {
		return Drop{}, err
	}
	return model(d, tags), nil
}

func Get(ctx context.Context, q db.Queryable, user api.User, id uuid.UUID) (Drop, error) {
	d, err := q.DropFind(ctx, db.DropFindParams{
		UserID: user.ID,
		ID:     id,
	})
	if err != nil {
		return Drop{}, err
	}
	return loadOne(ctx, q, user, d)
}

func Next(ctx context.Context, q db.Queryable, user api.User) (Drop, error) {
	d, err := q.DropNext(ctx, user.ID)
	if err != nil {
		return Drop{}, err
	}
	return loadOne(ctx, q, user, d)
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
	return loadMany(ctx, q, user, ds)
}

func Search(ctx context.Context, q db.Queryable, user api.User, body ListBody) ([]Drop, error) {
	qq := db.Pq.
		Select("drops").
		Join("drop_tags ON drop_tags.drop_id = drops.id").
		Join("tags ON tags.id = drop_tags.tag_id").
		Where(sq.Eq{
			"drops.user_id": user.ID,
			"tags.user_id":  user.ID,
		}).
		Limit(uint64(body.Limit))

	if body.Status != StatusUnknown {
		qq = qq.Where(sq.Eq{"drops.status": body.Status})
	}
	if body.Tags != nil {
		qq = qq.Where(sq.Eq{"tags.id": body.Tags})
	}

	query, args, err := qq.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	var drops []db.Drop
	if err := scan.RowsStrict(&drops, rows); err != nil {
		return nil, err
	}

	return loadMany(ctx, q, user, drops)
}

func loadOne(ctx context.Context, q db.Queryable, user api.User, d db.Drop) (Drop, error) {
	ts, err := q.TagsDrop(ctx, db.TagsDropParams{
		UserID: user.ID,
		ID:     d.ID,
	})
	return model(d, ts), err
}

func loadMany(ctx context.Context, q db.Queryable, user api.User, ds []db.Drop) ([]Drop, error) {
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
