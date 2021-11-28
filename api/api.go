package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/metagram-net/firehose/apierror"
	"github.com/metagram-net/firehose/auth"
	"github.com/metagram-net/firehose/clock"
	"github.com/metagram-net/firehose/db"
	"go.uber.org/zap"
)

var (
	ErrUnhandled = apierror.Error{
		Status:  http.StatusInternalServerError,
		Code:    "internal_server_error",
		Message: "Oops, sorry! There's an unhandled error in here somewhere.",
	}
	ErrNotFound = apierror.Error{
		Status:  http.StatusNotFound,
		Code:    "not_found",
		Message: "The requested route or resource does not exist.",
	}
)

type Context struct {
	Ctx   context.Context
	Log   *zap.Logger
	Tx    db.Queryable
	Clock clock.Clock
}

type HandlerFunc func(Context, auth.User, http.ResponseWriter, *http.Request)

type Server struct {
	log *zap.Logger
	db  *sql.DB
}

func NewServer(log *zap.Logger, db *sql.DB) *Server {
	return &Server{log, db}
}

func (s *Server) Authed(next HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			Respond(s.log, w, nil, err)
			return
		}
		defer func() {
			if err := tx.Commit(); err != nil {
				s.log.Error("Could not commit transaction", zap.Error(err))
			}
		}()

		u, err := auth.FromRequest(ctx, s.log, tx, r)
		if err != nil {
			Respond(s.log, w, nil, err)
			return
		}

		a := Context{
			Ctx:   ctx,
			Log:   s.log,
			Tx:    tx,
			Clock: clock.Freeze(time.Now()),
		}
		next(a, *u, w, r)
	}
}

func NewLogger() (*zap.Logger, error) {
	// TODO(prod): if production, zap.NewProduction()
	return zap.NewDevelopment()
}

func NewLogMiddleware(log *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info("Incoming request", zap.String("method", r.Method), zap.String("uri", r.RequestURI))
			next.ServeHTTP(w, r)
		})
	}
}

func WriteError(log *zap.Logger, w http.ResponseWriter, err error) error {
	var e apierror.Error
	if !errors.As(err, &e) {
		log.Warn("Unhandled error", zap.Error(err))
		e = ErrUnhandled
	}

	w.WriteHeader(e.Status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]string{
		"error_code":    string(e.Code),
		"error_message": e.Message,
	})
}

func WriteResult(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func Respond(log *zap.Logger, w http.ResponseWriter, v interface{}, err error) {
	var werr error
	if err == nil {
		werr = WriteResult(w, v)
	} else {
		werr = WriteError(log, w, err)
	}
	if werr != nil {
		log.Error("Could not write response, giving up", zap.Error(werr))
		panic(werr)
	}
}
