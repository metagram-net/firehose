package api

import (
	"context"

	"github.com/metagram-net/firehose/clock"
	"github.com/metagram-net/firehose/db"
	"go.uber.org/zap"
)

type Context struct {
	Ctx   context.Context
	Log   *zap.Logger
	Tx    db.Queryable
	Clock clock.Clock
}
