package drop

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func Register(r *mux.Router, db *sql.DB, log *zap.Logger) {
	r.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		d, err := Random(context.Background(), db)
		if err != nil {
			log.Error("Could not get random drop", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(d); err != nil {
			log.Error("Could not marshal drop", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
