package errs

import (
	"errors"
	"fmt"
)

// Error represents a business error.
type Error struct {
	Code    int    // business error code
	Message string // business error message
}

func New(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// AsError extracts the business error from the error chain.
func AsError(err error) (*Error, bool) {
	return errors.AsType[*Error](err)
}
