package server

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/auth"
	"github.com/metagram-net/firehose/drop"
	"github.com/metagram-net/firehose/wellknown"
	"go.uber.org/zap"
)

func New(log *zap.Logger, db *sql.DB) *mux.Router {
	srv := api.NewServer(log, db)
	handler := Handler{
		WellKnown: wellknown.Handler{},
		Auth:      auth.Handler{},
		Drops:     drop.Handler{},
	}

	router := Register(srv, handler)

	router.Use(api.NewLogMiddleware(log))

	router.NotFoundHandler = notFound(log)
	router.MethodNotAllowedHandler = methodNotAllowed(log)

	return router
}

func notFound(log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		api.Respond(log, w, nil, api.ErrNotFound)
	}
}

func methodNotAllowed(log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		api.Respond(log, w, nil, api.ErrMethodNotAllowed)
	}
}
