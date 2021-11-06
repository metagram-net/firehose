package drop

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func Register(r *mux.Router, db *sql.DB) {
	r.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		d, err := Random(context.Background(), db)
		if err != nil {
			log.Printf("could not get random drop: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(d); err != nil {
			log.Printf("could not marshal drop: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
