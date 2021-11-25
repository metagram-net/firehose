package api

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/metagram-net/firehose/clock"
	"github.com/metagram-net/firehose/db"
	"go.uber.org/zap"
)

type HandlerContext struct {
	Ctx   context.Context
	Log   *zap.Logger
	Tx    db.Queryable
	Clock clock.Clock
}

type Handler func(HandlerContext, http.ResponseWriter, *http.Request)

func Handle(db *sql.DB, log *zap.Logger, next Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			Respond(log, w, nil, err)
			return
		}
		defer func() {
			if err := tx.Commit(); err != nil {
				log.Error("Could not commit transaction", zap.Error(err))
			}
		}()

		hctx := HandlerContext{
			Ctx:   ctx,
			Log:   log,
			Tx:    tx,
			Clock: clock.Freeze(time.Now()),
		}
		next(hctx, w, r)
	})
}
