package drop_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/metagram-net/firehose/drop"
	"github.com/metagram-net/firehose/internal/apitest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	var (
		ctx   = apitest.Context(t, 500*time.Millisecond)
		clock = apitest.Clock(t)
		tx    = apitest.Tx(t, ctx)
		user  = apitest.User(t, ctx, tx)
	)
	title := "Example Dot Net"
	url := "https://example.net"

	d, err := drop.Create(ctx, tx, user, title, url, nil, clock.Now())
	require.NoError(t, err)

	assert.NoError(t, parseUUID(d.ID))
	assert.Equal(t, "Example Dot Net", d.Title)
	assert.Equal(t, "https://example.net", d.URL)
	assert.Equal(t, drop.StatusUnread, d.Status)
	assert.WithinDuration(t, clock.Now(), *d.MovedAt, 0)
	assert.Empty(t, d.Tags)
}

func parseUUID(s string) error {
	_, err := uuid.Parse(s)
	if err != nil {
		return fmt.Errorf(`parse UUID: "%s": %w`, s, err)
	}
	return nil
}
