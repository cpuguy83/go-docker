package errdefs

import (
	"errors"
	"fmt"
)

// Unavailable is an error interface which denotes whether the opration failed due
// to a the resource not being found.
type ErrUnavailable interface {
	Unavailable() bool
	error
}

type unavailableError struct {
	error
}

func (e *unavailableError) Unavailable() bool {
	return true
}

func (e *unavailableError) Cause() error {
	return e.error
}

// AsUnavailable wraps the passed in error to make it of type ErrUnavailable
//
// Callers should make sure the passed in error has exactly the error message
// it wants as this function does not decorate the message.
func AsUnavailable(err error) error {
	if err == nil {
		return nil
	}
	return &unavailableError{err}
}

// Unavailable makes an ErrUnavailable from the provided error message
func Unavailable(msg string) error {
	return &unavailableError{errors.New(msg)}
}

// Unavailablef makes an ErrUnavailable from the provided error format and args
func Unavailablef(format string, args ...interface{}) error {
	return &unavailableError{fmt.Errorf(format, args...)}
}

// IsUnavailable determines if the passed in error is of type ErrUnavailable
//
// This will traverse the causal chain (`Cause() error`), until it finds an error
// which implements the `Unavailable` interface.
func IsUnavailable(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(ErrUnavailable); ok {
		return e.Unavailable()
	}

	if e, ok := err.(causal); ok {
		return IsUnavailable(e.Cause())
	}

	return false
}
