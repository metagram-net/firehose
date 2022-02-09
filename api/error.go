package api

import (
	"fmt"
	"net/http"
)

type httpStatus = int
type ErrorCode string

type Error struct {
	Status  httpStatus `json:"status"`
	Code    ErrorCode  `json:"code"`
	Message string     `json:"message"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

var (
	ErrUnhandled = Error{
		Status:  http.StatusInternalServerError,
		Code:    "internal_server_error",
		Message: "Oops, sorry! There's an unhandled error in here somewhere.",
	}
	ErrNotFound = Error{
		Status:  http.StatusNotFound,
		Code:    "not_found",
		Message: "The requested route does not exist.",
	}
	ErrMethodNotAllowed = Error{
		Status: http.StatusMethodNotAllowed,
		Code:   "method_not_allowed",
		// TODO: Include the requested method and allowed methods in this message. (gorilla/mux #652)
		Message: "The requested HTTP method cannot be handled by this route.",
	}
	ErrMissingAuthz = Error{
		Status:  http.StatusUnauthorized,
		Code:    "missing_authorization",
		Message: "The Authorization header was missing or the wrong format.",
	}
	ErrInvalidAuthz = Error{
		Status:  http.StatusUnauthorized,
		Code:    "invalid_authorization",
		Message: "The provided credentials were not valid.",
	}
)

func ValidationError(field string, value interface{}, message string) Error {
	return Error{
		Status:  http.StatusBadRequest,
		Code:    "validation_error",
		Message: fmt.Sprintf(`field "%s" (value "%s"): %s`, field, value, message),
	}
}
