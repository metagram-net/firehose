package drop

import (
	"github.com/gofrs/uuid"

	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/db"
)

type ID uuid.UUID

func (id *ID) FromParam(s string) error {
	u, err := uuid.FromString(s)
	*id = ID(u)
	return err
}

func (id ID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

type Handler struct{}

func (Handler) Next(ctx api.Context, user api.User) (Drop, error) {
	q := db.New(ctx.Tx)
	return Next(ctx, q, user)
}

type GetParams struct {
	ID ID `var:"id"`
}

func (Handler) Get(ctx api.Context, u api.User, params GetParams) (Drop, error) {
	q := db.New(ctx.Tx)
	return Get(ctx, q, u, params.ID.UUID())
}

type ListBody struct {
	Status Status       `json:"status,omitempty"`
	Limit  *int32       `json:"limit,omitempty"`
	Tags   *[]uuid.UUID `json:"tags"`
}

type ListResponse struct {
	Drops []Drop `json:"drops"`
}

func (Handler) List(ctx api.Context, u api.User, body ListBody) (ListResponse, error) {
	// The point of Firehose is to force me to consume or delete the oldest
	// content first. Listing is useful if some content is not currently
	// consumable (videos and PDFs are the usual examples) so you have to
	// "scroll past" a few drops. But the point is to avoid scrolling for a
	// long time to find something, so don't allow large limits here.
	limit := int32(20)
	l := body.Limit
	if l != nil && 0 < *l && *l < limit {
		limit = *l
	}
	body.Limit = &limit

	q := db.New(ctx.Tx)

	// The query is simpler (and probably faster?) if it doesn't have to join
	// tables to filter on tags.
	if body.Tags == nil {
		ds, err := List(ctx, q, u, body.Status, limit)
		return ListResponse{Drops: ds}, err
	}

	// TODO: Should search be its own endpoint?
	ds, err := Search(ctx, q, u, body)
	return ListResponse{Drops: ds}, err
}

type CreateBody struct {
	Title  string      `json:"title,omitempty"`
	URL    string      `json:"url,omitempty"`
	TagIDs []uuid.UUID `json:"tag_ids,omitempty"`
}

func (Handler) Create(ctx api.Context, u api.User, body CreateBody) (Drop, error) {
	q := db.New(ctx.Tx)
	now := ctx.Clock.Now()
	return Create(ctx, q, u, body.Title, body.URL, body.TagIDs, now)
}

type UpdateBody struct {
	ID    uuid.UUID    `json:"id,omitempty"`
	Title *string      `json:"title,omitempty"`
	URL   *string      `json:"url,omitempty"`
	Tags  *[]uuid.UUID `json:"tags,omitempty"`
}

func (Handler) Update(ctx api.Context, u api.User, body UpdateBody) (Drop, error) {
	q := db.New(ctx.Tx)
	return Update(ctx, q, u, body.ID, UpdateFields{
		Title: body.Title,
		URL:   body.URL,
		Tags:  body.Tags,
	})
}

type MoveBody struct {
	ID     uuid.UUID `json:"id,omitempty"`
	Status Status    `json:"status,omitempty"`
}

func (Handler) Move(ctx api.Context, u api.User, body MoveBody) (Drop, error) {
	q := db.New(ctx.Tx)
	now := ctx.Clock.Now()
	return Move(ctx, q, u, body.ID, body.Status, now)
}

type DeleteBody struct {
	ID uuid.UUID `json:"id,omitempty"`
}

func (Handler) Delete(ctx api.Context, u api.User, body DeleteBody) (Drop, error) {
	q := db.New(ctx.Tx)
	return Delete(ctx, q, u, body.ID)
}
