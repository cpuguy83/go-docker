package errdefs

import (
	"errors"
	"fmt"
)

var ErrConflict = errors.New("conflict")

// Conflict makes an ErrConflict from the provided error message
func Conflict(msg string) error {
	return fmt.Errorf("%w: %s", ErrConflict, msg)
}

// Conflictf makes an ErrConflict from the provided error format and args
func Conflictf(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrConflict, fmt.Sprintf(format, args...))
}

// IsConflict determines if the passed in error is of type ErrConflict
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// AsConflict returns a wrapped error which will return true for IsConflict
func AsConflict(err error) error {
	return as(err, ErrConflict)
}
