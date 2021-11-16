package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

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
