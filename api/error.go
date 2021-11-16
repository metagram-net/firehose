package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type httpStatus = int
type ErrorCode string

type Error struct {
	Status  httpStatus
	Code    ErrorCode
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(status httpStatus, code ErrorCode, message string) Error {
	return Error{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

var ErrUnhandled = Error{
	Status:  http.StatusInternalServerError,
	Code:    "internal_server_error",
	Message: "Oops, sorry! There's an unhandled error in here somewhere.",
}

func WriteError(log *zap.Logger, w http.ResponseWriter, err error) error {
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
