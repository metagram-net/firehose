package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/metagram-net/firehose/auth/apikey"
	"github.com/metagram-net/firehose/auth/user"
	"github.com/metagram-net/firehose/clock"
	"github.com/metagram-net/firehose/db"
	"go.uber.org/zap"
)

type User struct {
	ID uuid.UUID `json:"id"`
}

type Context struct {
	context.Context

	Log   *zap.Logger
	Tx    *sql.Tx
	Clock clock.Clock
}

func (c *Context) Close() error {
	return c.Tx.Commit()
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

		q := db.New(tx)
		u, err := authenticate(ctx, s.log, q, r)
		if err != nil {
			Respond(s.log, w, nil, err)
			return
		}

		a := Context{
			Context: ctx,
			Log:     s.log,
			Tx:      tx,
			Clock:   clock.Freeze(time.Now()),
		}
		next(a, *u, w, r)
	}
}

func (s *Server) Context(r *http.Request) (Context, error) {
	var zero Context

	ctx := r.Context()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}

	return Context{
		Context: ctx,
		Log:     s.log,
		Tx:      tx,
		Clock:   clock.Freeze(time.Now()),
	}, nil
}

func (s *Server) Authenticate(ctx Context, r *http.Request) (*User, error) {
	q := db.New(ctx.Tx)
	return authenticate(ctx, s.log, q, r)
}

func (s *Server) Respond(w http.ResponseWriter, v interface{}, err error) {
	Respond(s.log, w, v, err)
}

// TODO: Replace API key "passwords" with PK registration and request signing.

func authenticate(ctx context.Context, log *zap.Logger, q db.Querier, req *http.Request) (*User, error) {
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

	key, err := apikey.Find(ctx, q, userID, token)
	if err != nil {
		log.Warn("Could not find API key by user ID and token", zap.Error(err))
		return nil, ErrInvalidAuthz
	}
	u, err := user.Find(ctx, q, key.UserID)
	if err != nil {
		log.Warn("Could not find user by ID", zap.Error(err))
		return nil, ErrInvalidAuthz
	}
	return &User{ID: u.ID}, nil
}

func NewLogMiddleware(log *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info("Incoming request", zap.String("method", r.Method), zap.String("uri", r.RequestURI))
			next.ServeHTTP(w, r)
		})
	}
}

func writeError(log *zap.Logger, w http.ResponseWriter, err error) error {
	var e Error
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

func writeResult(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

// TODO: Deprecate top-level Respond in favor of Server.Respond

func Respond(log *zap.Logger, w http.ResponseWriter, v interface{}, err error) {
	var werr error
	if err == nil {
		werr = writeResult(w, v)
	} else {
		werr = writeError(log, w, err)
	}
	if werr != nil {
		log.Error("Could not write response, giving up", zap.Error(werr))
		panic(werr)
	}
}

type Validator interface {
	Validate() error
}

// Parse reads the request body and unmarshals the JSON into v. For types that
// implement Validator, this returns v.Validate() after unmarshaling.
func Parse(r *http.Request, v interface{}) error {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, v)
	if err != nil {
		return err
	}

	if vv, ok := v.(Validator); ok {
		return vv.Validate()
	}
	return nil
}

type Param interface {
	FromParam(string) error
}

// Parse unpacks URL parameters into v. For types that implement Validator,
// this returns v.Validate() after unpacking.
//
// This is intended to be called with the result of mux.Vars.
//
// Unpacking is done with json.Marshal, with some overrides for specific types:
//
// - If a type implements Param, its FromParam method will be used.
// - For string and []byte, the value will be returned as-is.
// - Complex numbers will use strconv.ParseComplex with the correct bit size.
//
func FromVars(vars map[string]string, v interface{}) error {
	// v must be a pointer type for this to be able to set values.
	vv := reflect.ValueOf(v).Elem()
	t := vv.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		key := f.Tag.Get("var")
		if key == "" {
			continue
		}

		dest := vv.FieldByIndex(f.Index).Addr().Interface()
		err := fromVar(vars[key], dest)
		if err != nil {
			return fmt.Errorf("decode param: %w", err)
		}
	}

	if vv, ok := v.(Validator); ok {
		return vv.Validate()
	}
	return nil
}

var ErrFromVar = errors.New("fromVar: unhandled type")

func fromVar(val string, dest interface{}) error {
	var err error
	switch dest := dest.(type) {
	// If the type implements its own decoding, use that.
	case Param:
		err = dest.FromParam(val)

	// There's no need to decode these cases. They'd fail in JSON anyway
	// because these values aren't quoted.
	case *string:
		*dest = val
	case *[]byte:
		*dest = []byte(val)

	// Complex numbers aren't supported by JSON.
	case *complex64:
		f, err := strconv.ParseComplex(val, 64)
		*dest = complex64(f)
		return err
	case *complex128:
		*dest, err = strconv.ParseComplex(val, 128)

	default:
		err = json.Unmarshal([]byte(val), dest)
	}
	return err
}

// FromBody reads the request body and unmarshals the JSON into v. For types
// that implement Validator, this returns v.Validate() after unmarshaling.
func FromBody(body io.Reader, v interface{}) error {
	err := json.NewDecoder(body).Decode(v)
	if err != nil {
		return err
	}

	if vv, ok := v.(Validator); ok {
		return vv.Validate()
	}
	return nil
}
