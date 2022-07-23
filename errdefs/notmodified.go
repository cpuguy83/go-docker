package errdefs

import (
	"errors"
	"fmt"
)

var ErrNotModified = errors.New("not modified")

// NotModified makes an ErrNotModified from the provided error message
func NotModified(msg string) error {
	return fmt.Errorf("%w: %s", ErrNotModified, msg)
}

// NotModifiedf makes an ErrNotModified from the provided error format and args
func NotModifiedf(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrNotModified, fmt.Sprintf(format, args...))
}

// IsNotModified determines if the passed in error is of type ErrNotModified
func IsNotModified(err error) bool {
	return errors.Is(err, ErrNotModified)
}

// AsNotModified returns a wrapped error which will return true for IsNotModified
func AsNotModified(err error) error {
	return as(err, ErrNotModified)
}
