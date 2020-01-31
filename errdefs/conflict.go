package errdefs

import (
	"errors"
	"fmt"
)

// Conflict is an error interface which denotes whether the opration failed due
// to a the resource not being found.
type ErrConflict interface {
	Conflict() bool
	error
}

type conflictError struct {
	error
}

func (e *conflictError) Conflict() bool {
	return true
}

func (e *conflictError) Cause() error {
	return e.error
}

// AsConflict wraps the passed in error to make it of type ErrConflict
//
// Callers should make sure the passed in error has exactly the error message
// it wants as this function does not decorate the message.
func AsConflict(err error) error {
	if err == nil {
		return nil
	}
	return &conflictError{err}
}

// Conflict makes an ErrConflict from the provided error message
func Conflict(msg string) error {
	return &conflictError{errors.New(msg)}
}

// Conflictf makes an ErrConflict from the provided error format and args
func Conflictf(format string, args ...interface{}) error {
	return &conflictError{fmt.Errorf(format, args...)}
}

// IsConflict determines if the passed in error is of type ErrConflict
//
// This will traverse the causal chain (`Cause() error`), until it finds an error
// which implements the `Conflict` interface.
func IsConflict(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(ErrConflict); ok {
		return e.Conflict()
	}

	if e, ok := err.(causal); ok {
		return IsConflict(e.Cause())
	}

	return false
}
