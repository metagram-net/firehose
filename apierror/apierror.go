package apierror

import (
	"fmt"
)

// TODO: Move these types to api (remove circular dependency with auth first)

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
