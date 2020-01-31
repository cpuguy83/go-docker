package errdefs

import (
	"errors"
	"fmt"
)

// NotModified is an error interface which denotes whether the opration failed due
// to a the resource not being found.
type ErrNotModified interface {
	NotModified() bool
	error
}

type notModifiedError struct {
	error
}

func (e *notModifiedError) NotModified() bool {
	return true
}

func (e *notModifiedError) Cause() error {
	return e.error
}

// AsNotModified wraps the passed in error to make it of type ErrNotModified
//
// Callers should make sure the passed in error has exactly the error message
// it wants as this function does not decorate the message.
func AsNotModified(err error) error {
	if err == nil {
		return nil
	}
	return &notModifiedError{err}
}

// NotModified makes an ErrNotModified from the provided error message
func NotModified(msg string) error {
	return &notModifiedError{errors.New(msg)}
}

// NotModifiedf makes an ErrNotModified from the provided error format and args
func NotModifiedf(format string, args ...interface{}) error {
	return &notModifiedError{fmt.Errorf(format, args...)}
}

// IsNotModified determines if the passed in error is of type ErrNotModified
//
// This will traverse the causal chain (`Cause() error`), until it finds an error
// which implements the `NotModified` interface.
func IsNotModified(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(ErrNotModified); ok {
		return e.NotModified()
	}

	if e, ok := err.(causal); ok {
		return IsNotModified(e.Cause())
	}

	return false
}
