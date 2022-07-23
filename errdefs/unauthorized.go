package errdefs

import (
	"errors"
	"fmt"
)

var ErrUnauthorized = errors.New("unauthorized")

// Unauthorized makes an ErrUnauthorized from the provided error message
func Unauthorized(msg string) error {
	return fmt.Errorf("%w: %s", ErrUnauthorized, msg)
}

// Unauthorizedf makes an ErrUnauthorized from the provided error format and args
func Unauthorizedf(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrUnauthorized, fmt.Sprintf(format, args...))
}

// IsUnauthorized determines if the passed in error is of type ErrUnauthorized
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsUnauthorized determines if the passed in error is of type ErrUnauthorized
func AsUnauthorized(err error) error {
	return as(err, ErrUnauthorized)
}
