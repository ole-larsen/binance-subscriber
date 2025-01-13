package poller

import (
	"fmt"
)

var ErrConnectionNotInitialized = NewError(fmt.Errorf("connection is not initialized"))

// Error - custom client error.
type Error struct {
	err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("[poller]: %v", e.err)
}

func NewError(err error) error {
	if err == nil {
		return nil
	}

	return &Error{
		err: err,
	}
}

func (e *Error) Unwrap() error {
	return e.err
}
