package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/metagram-net/firehose/apierror"
	"github.com/metagram-net/firehose/clock"
	"github.com/metagram-net/firehose/db"
	"go.uber.org/zap"
)

type Context struct {
	Ctx   context.Context
	Log   *zap.Logger
	Tx    db.Queryable
	Clock clock.Clock
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
		e = apierror.ErrUnhandled
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
