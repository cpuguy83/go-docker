package errdefs

import (
	"errors"
	"fmt"
)

// Unauthorized is an error interface which denotes whether the opration failed due
// to a the resource not being found.
type ErrUnauthorized interface {
	Unauthorized() bool
	error
}

type unauthorizedError struct {
	error
}

func (e *unauthorizedError) Unauthorized() bool {
	return true
}

func (e *unauthorizedError) Cause() error {
	return e.error
}

// AsUnauthorized wraps the passed in error to make it of type ErrUnauthorized
//
// Callers should make sure the passed in error has exactly the error message
// it wants as this function does not decorate the message.
func AsUnauthorized(err error) error {
	if err == nil {
		return nil
	}
	return &unauthorizedError{err}
}

// Unauthorized makes an ErrUnauthorized from the provided error message
func Unauthorized(msg string) error {
	return &unauthorizedError{errors.New(msg)}
}

// Unauthorizedf makes an ErrUnauthorized from the provided error format and args
func Unauthorizedf(format string, args ...interface{}) error {
	return &unauthorizedError{fmt.Errorf(format, args...)}
}

// IsUnauthorized determines if the passed in error is of type ErrUnauthorized
//
// This will traverse the causal chain (`Cause() error`), until it finds an error
// which implements the `Unauthorized` interface.
func IsUnauthorized(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(ErrUnauthorized); ok {
		return e.Unauthorized()
	}

	if e, ok := err.(causal); ok {
		return IsUnauthorized(e.Cause())
	}

	return false
}
