// error.go defined our own error type
package sessions

import (
	"fmt"
)

type Error *internalError

// internalError encompasses various kinds of error states we can have
type internalError struct {
	// Technical error that happend (e.g. query failed)
	Err error
	// Message to send to the user
	Message string
}

func NewError(message string, err error) Error {
	return &internalError{Message: message, Err: err}
}

func PassThrough(err error) Error {
	if err == nil {
		return nil
	}
	return &internalError{Err: err}
}

func ToError(e Error) error {
	return fmt.Errorf("%s: %v", e.Message, e.Err)
}
