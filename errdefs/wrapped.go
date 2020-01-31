package errdefs

// This pattern is used by github.com/pkg/errors
type causal interface {
	Cause() error
	error
}
