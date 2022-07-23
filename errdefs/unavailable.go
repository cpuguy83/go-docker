package errdefs

import (
	"errors"
	"fmt"
)

var ErrUnavailable = errors.New("unavailable")

// Unavailable makes an ErrUnavailable from the provided error message
func Unavailable(msg string) error {
	return fmt.Errorf("%w: %s", ErrUnavailable, msg)
}

// Unavailablef makes an ErrUnavailable from the provided error format and args
func Unavailablef(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrUnavailable, fmt.Sprintf(format, args...))
}

// IsUnavailable determines if the passed in error is of type ErrUnavailable
func IsUnavailable(err error) bool {
	return errors.Is(err, ErrUnavailable)
}

// AsUnavailable returns a wrapped error which will return true for IsUnavailable
func AsUnavailable(err error) error {
	return as(err, ErrUnavailable)
}
