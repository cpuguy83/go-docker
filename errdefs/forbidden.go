package errdefs

import (
	"errors"
	"fmt"
)

var ErrForbidden = errors.New("forbidden")

// Forbidden makes an ErrForbidden from the provided error message
func Forbidden(msg string) error {
	return fmt.Errorf("%w: %s", ErrForbidden, msg)
}

// Forbiddenf makes an ErrForbidden from the provided error format and args
func Forbiddenf(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrForbidden, fmt.Sprintf(format, args...))
}

// IsForbidden determines if the passed in error is of type ErrForbidden
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// AsForbidden returns a wrapped error which will return true for IsForbidden
func AsForbidden(err error) error {
	return as(err, ErrForbidden)
}
