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

	// TODO: Maybe it makes sense to move the *.Server structs to here?

	r.Methods(http.MethodGet).Path("/.well-known/health-check").HandlerFunc(wellknown.HealthCheck)

	authSrv := auth.NewServer(log, db)
	r.Methods(http.MethodGet).Path("/auth/whoami").HandlerFunc(authSrv.Whoami)

	dropSrv := drop.NewServer(log, db)
	r.Methods(http.MethodGet).Path("/v1/drops/random").HandlerFunc(dropSrv.Random)
	r.Methods(http.MethodGet).Path("/v1/drops/next").HandlerFunc(dropSrv.Next)
	r.Methods(http.MethodGet).Path("/v1/drops/get/{id}").HandlerFunc(dropSrv.Get)
	r.Methods(http.MethodPost).Path("/v1/drops/create").HandlerFunc(dropSrv.Create)
	r.Methods(http.MethodPost).Path("/v1/drops/update/{id}").HandlerFunc(dropSrv.Update)
	r.Methods(http.MethodPost).Path("/v1/drops/delete/{id}").HandlerFunc(dropSrv.Delete)

	// TODO: mux.Router.NotFoundHandler
	// TODO: mux.Router.MethodNotAllowedHandler

	return r
}
