package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/metagram-net/firehose/apierror"
	"github.com/metagram-net/firehose/auth/apikey"
	"github.com/metagram-net/firehose/auth/user"
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
	ErrMethodNotAllowed = apierror.Error{
		Status: http.StatusMethodNotAllowed,
		Code:   "method_not_allowed",
		// TODO: Include the requested method and allowed methods in this message. (gorilla/mux #652)
		Message: "The requested HTTP method cannot be handled by this route.",
	}
	ErrMissingAuthz = apierror.Error{
		Status:  http.StatusUnauthorized,
		Code:    "missing_authorization",
		Message: "The Authorization header was missing or the wrong format.",
	}
	ErrInvalidAuthz = apierror.Error{
		Status:  http.StatusUnauthorized,
		Code:    "invalid_authorization",
		Message: "The provided credentials were not valid.",
	}
)

type User struct {
	ID uuid.UUID `json:"id"`
}

type Context struct {
	Ctx   context.Context
	Log   *zap.Logger
	Tx    db.Queryable
	Clock clock.Clock
}

type HandlerFunc func(Context, User, http.ResponseWriter, *http.Request)

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

		u, err := authenticate(ctx, s.log, tx, r)
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

// TODO: Replace API key "passwords" with PK registration and request signing.

func authenticate(ctx context.Context, log *zap.Logger, tx db.Queryable, req *http.Request) (*User, error) {
	// If the header is missing completely, return a more helpful error.
	if req.Header.Get("Authorization") == "" {
		return nil, ErrMissingAuthz
	}

	// Parse out the user ID and token
	username, password, ok := req.BasicAuth()
	if !ok {
		log.Warn("Invalid basic auth header")
		return nil, ErrInvalidAuthz
	}
	userID, err := uuid.FromString(username)
	if err != nil {
		log.Warn("Could not parse user ID", zap.Error(err))
		return nil, ErrInvalidAuthz
	}
	token, err := apikey.NewPlaintext(password)
	if err != nil {
		log.Warn("Could not parse token", zap.Error(err))
		return nil, ErrInvalidAuthz
	}

	key, err := apikey.Find(ctx, tx, userID, token)
	if err != nil {
		log.Warn("Could not find API key by user ID and token", zap.Error(err))
		return nil, ErrInvalidAuthz
	}
	u, err := user.Find(ctx, tx, key.UserID)
	if err != nil {
		log.Warn("Could not find user by ID", zap.Error(err))
		return nil, ErrInvalidAuthz
	}
	return &User{ID: u.ID}, nil
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
