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

	router.NotFoundHandler = notFound(srv)
	router.MethodNotAllowedHandler = methodNotAllowed(srv)

	return router
}

func notFound(srv *api.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		srv.Respond(w, nil, api.ErrNotFound)
	}
}

func methodNotAllowed(srv *api.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		srv.Respond(w, nil, api.ErrMethodNotAllowed)
	}
}
