package firehose

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

func Server(log *zap.Logger, db *sql.DB) *mux.Router {
	r := mux.NewRouter()

	r.Use(api.NewLogMiddleware(log))

	srv := api.NewServer(log, db)

	r.Methods(http.MethodGet).Path("/.well-known/health-check").HandlerFunc(wellknown.HealthCheck)

	r.Methods(http.MethodGet).Path("/auth/whoami").HandlerFunc(srv.Authed(whoami))

	var dropSrv drop.Handler
	r.Methods(http.MethodGet).Path("/v1/drops/random").HandlerFunc(srv.Authed(dropSrv.Random))
	r.Methods(http.MethodGet).Path("/v1/drops/next").HandlerFunc(srv.Authed(dropSrv.Next))
	r.Methods(http.MethodGet).Path("/v1/drops/get/{id}").HandlerFunc(srv.Authed(dropSrv.Get))
	r.Methods(http.MethodPost).Path("/v1/drops/create").HandlerFunc(srv.Authed(dropSrv.Create))
	r.Methods(http.MethodPost).Path("/v1/drops/update/{id}").HandlerFunc(srv.Authed(dropSrv.Update))
	r.Methods(http.MethodPost).Path("/v1/drops/delete/{id}").HandlerFunc(srv.Authed(dropSrv.Delete))

	r.NotFoundHandler = notFound(log)
	// TODO: mux.Router.MethodNotAllowedHandler

	return r
}

func whoami(a api.Context, u auth.User, w http.ResponseWriter, r *http.Request) {
	api.Respond(a.Log, w, u, nil)
}

func notFound(log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		api.Respond(log, w, nil, api.ErrNotFound)
	}
}
