package errdefs

import (
	"errors"
	"fmt"
)

// NotImplemented is an error interface which denotes whether the opration failed due
// to a the resource not being found.
type ErrNotImplemented interface {
	NotImplemented() bool
	error
}

type notImplementedError struct {
	error
}

func (e *notImplementedError) NotImplemented() bool {
	return true
}

func (e *notImplementedError) Cause() error {
	return e.error
}

// AsNotImplemented wraps the passed in error to make it of type ErrNotImplemented
//
// Callers should make sure the passed in error has exactly the error message
// it wants as this function does not decorate the message.
func AsNotImplemented(err error) error {
	if err == nil {
		return nil
	}
	return &notImplementedError{err}
}

// NotImplemented makes an ErrNotImplemented from the provided error message
func NotImplemented(msg string) error {
	return &notImplementedError{errors.New(msg)}
}

// NotImplementedf makes an ErrNotImplemented from the provided error format and args
func NotImplementedf(format string, args ...interface{}) error {
	return &notImplementedError{fmt.Errorf(format, args...)}
}

// IsNotImplemented determines if the passed in error is of type ErrNotImplemented
//
// This will traverse the causal chain (`Cause() error`), until it finds an error
// which implements the `NotImplemented` interface.
func IsNotImplemented(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(ErrNotImplemented); ok {
		return e.NotImplemented()
	}

	if e, ok := err.(causal); ok {
		return IsNotImplemented(e.Cause())
	}

	return false
}
