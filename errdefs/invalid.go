package errdefs

import (
	"errors"
	"fmt"
)

var ErrInvalid = errors.New("invalid")

// InvalidInput makes an ErrInvalidInput from the provided error message
func Invalid(msg string) error {
	return fmt.Errorf("%w: %s", ErrInvalid, msg)
}

// InvalidInputf makes an ErrInvalidInput from the provided error format and args
func Invalidf(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

// IsInvalid determines if the passed in error is of type ErrIsInvalid
func IsInvalid(err error) bool {
	return errors.Is(err, ErrInvalid)
}

// AsInvalid returns a wrapped error which will return true for IsInvalid
func AsInvalid(err error) error {
	return as(err, ErrInvalid)
}
