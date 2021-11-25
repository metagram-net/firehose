package apierror

import (
	"fmt"
	"net/http"
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

func New(status httpStatus, code ErrorCode, message string) Error {
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
