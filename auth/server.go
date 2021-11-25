package auth

import (
	"database/sql"
	"net/http"

	"github.com/metagram-net/firehose/api"
	"go.uber.org/zap"
)

type Server struct {
	log *zap.Logger
	db  *sql.DB
}

func NewServer(log *zap.Logger, db *sql.DB) *Server {
	return &Server{log, db}
}

func (s *Server) Whoami(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		api.Respond(s.log, w, nil, err)
		return
	}
	defer func() {
		if err := tx.Commit(); err != nil {
			s.log.Error("Could not commit transaction", zap.Error(err))
		}
	}()

	u, err := Whoami(ctx, s.log, tx, r)
	api.Respond(s.log, w, u, err)
}
