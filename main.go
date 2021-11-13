package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib" // database/sql driver: pgx
	"github.com/metagram-net/firehose/auth"
	"github.com/metagram-net/firehose/drop"
	"github.com/metagram-net/firehose/wellknown"
)

func main() {
	if err := run(); err != nil {
		// TODO: graceful shutdown
		log.Fatal(err)
	}
	log.Print("Clean shutdown. Bye! 👋")
}

func run() error {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer db.Close()

	r := mux.NewRouter()
	wellknown.Register(r.PathPrefix("/.well-known/").Subrouter())
	auth.Register(r.PathPrefix("/auth/").Subrouter(), db)
	drop.Register(r.PathPrefix("/v1/drops/").Subrouter(), db)

	port := 8002
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Listening on %s\n", addr)
	return http.ListenAndServe(addr, r)
}
