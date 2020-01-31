package errdefs

import (
	"errors"
	"fmt"
)

// Forbidden is an error interface which denotes whether the opration failed due
// to a the resource not being found.
type ErrForbidden interface {
	Forbidden() bool
	error
}

type forbiddenError struct {
	error
}

func (e *forbiddenError) Forbidden() bool {
	return true
}

func (e *forbiddenError) Cause() error {
	return e.error
}

// AsForbidden wraps the passed in error to make it of type ErrForbidden
//
// Callers should make sure the passed in error has exactly the error message
// it wants as this function does not decorate the message.
func AsForbidden(err error) error {
	if err == nil {
		return nil
	}
	return &forbiddenError{err}
}

// Forbidden makes an ErrForbidden from the provided error message
func Forbidden(msg string) error {
	return &forbiddenError{errors.New(msg)}
}

// Forbiddenf makes an ErrForbidden from the provided error format and args
func Forbiddenf(format string, args ...interface{}) error {
	return &forbiddenError{fmt.Errorf(format, args...)}
}

// IsForbidden determines if the passed in error is of type ErrForbidden
//
// This will traverse the causal chain (`Cause() error`), until it finds an error
// which implements the `Forbidden` interface.
func IsForbidden(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(ErrForbidden); ok {
		return e.Forbidden()
	}

	if e, ok := err.(causal); ok {
		return IsForbidden(e.Cause())
	}

	return false
}
