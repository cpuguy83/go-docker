package errdefs

import (
	"errors"
	"fmt"
)

func as(e, target error) error {
	return &wrapped{e: e, cause: target}
}

type wrapped struct {
	cause error
	e     error
}

func (w *wrapped) Error() string {
	return fmt.Sprintf("%s: %s", w.cause.Error(), w.e.Error())
}

func (w *wrapped) Unwrap() error {
	return w.cause
}

func (w *wrapped) Is(target error) bool {
	if errors.Is(w.e, target) {
		return true
	}
	if errors.Is(w.cause, target) {
		return true
	}
	return false
}

// Wrap is a convenience function to wrap an error with extra text.
func Wrap(err error, msg string) error {
	return fmt.Errorf("%w: %s", err, msg)
}

// Wrapf is a convenience function to wrap an error with extra text.
func Wrapf(err error, format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", err, fmt.Sprintf(format, args...))
}
