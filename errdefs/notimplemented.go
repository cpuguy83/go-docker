package errdefs

import (
	"errors"
	"fmt"
)

var ErrNotImplemented = errors.New("not implemented")

// NotImplemented makes an ErrNotImplemented from the provided error message
func NotImplemented(msg string) error {
	return fmt.Errorf("%w: %s", ErrNotImplemented, msg)
}

// NotImplementedf makes an ErrNotImplemented from the provided error format and args
func NotImplementedf(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrNotImplemented, fmt.Sprintf(format, args...))
}

// IsNotImplemented determines if the passed in error is of type ErrNotImplemented
func IsNotImplemented(err error) bool {
	return errors.Is(err, ErrNotImplemented)
}

// AsNotImplemented returns a wrapped error which will return true for IsNotImplemented
func AsNotImplemented(err error) error {
	return as(err, ErrNotImplemented)
}
