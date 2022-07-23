package errdefs

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

// NotFound makes an ErrNotFound from the provided error message
func NotFound(msg string) error {
	return fmt.Errorf("%w: %s", ErrNotFound, msg)
}

// NotFoundf makes an ErrNotFound from the provided error format and args
func NotFoundf(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrNotFound, fmt.Sprintf(format, args...))
}

// IsNotFound determines if the passed in error is of type ErrNotFound
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// AsNotFound returns a wrapped error which will return true for IsNotFound
func AsNotFound(err error) error {
	return as(err, ErrNotFound)
}
