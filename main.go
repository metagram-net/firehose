package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib" // database/sql driver: pgx
	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/auth"
	"github.com/metagram-net/firehose/drop"
	"github.com/metagram-net/firehose/wellknown"
	"go.uber.org/zap"
)

func main() {
	log, err := api.NewLogger()
	if err != nil {
		panic(err)
	}

	log.Info("Starting database connection pool")
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	srv := server(log, db)

	// TODO(prod): Take host:port from config/env
	port := 8002
	addr := fmt.Sprintf(":%d", port)
	log.Info("Listening", zap.String("address", addr))
	if err := http.ListenAndServe(addr, srv); err != nil {
		// TODO(prod): graceful shutdown
		log.Fatal("Error during shutdown", zap.Error(err))
	}

	log.Info("Clean shutdown. Bye! 👋")
	if err := log.Sync(); err != nil {
		panic(err)
	}
}

func server(log *zap.Logger, db *sql.DB) *mux.Router {
	r := mux.NewRouter()

	r.Use(api.NewLogMiddleware(log))

	wellknown.Register(r.PathPrefix("/.well-known/").Subrouter())
	auth.Register(r.PathPrefix("/auth/").Subrouter(), db, log)
	drop.Register(r.PathPrefix("/v1/drops/").Subrouter(), db, log)

	// TODO: mux.Router.NotFoundHandler
	// TODO: mux.Router.MethodNotAllowedHandler

	return r
}
