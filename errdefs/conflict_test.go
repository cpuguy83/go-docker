package errdefs

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

type testingConflictError bool

func (e testingConflictError) Error() string {
	return fmt.Sprintf("%v", bool(e))
}

func (e testingConflictError) Conflict() bool {
	return bool(e)
}

func TestIsConflict(t *testing.T) {
	type testCase struct {
		name      string
		err       error
		xMsg      string
		xConflict bool
	}

	for _, c := range []testCase{
		{
			name:      "Conflictf",
			err:       Conflictf("%s not found", "foo"),
			xMsg:      "foo not found",
			xConflict: true,
		},
		{
			name:      "AsConflict",
			err:       AsConflict(errors.New("this is a test")),
			xMsg:      "this is a test",
			xConflict: true,
		},
		{
			name:      "AsConflictWithNil",
			err:       AsConflict(nil),
			xMsg:      "",
			xConflict: false,
		},
		{
			name:      "nilError",
			err:       nil,
			xMsg:      "",
			xConflict: false,
		},
		{
			name:      "customConflictFalse",
			err:       testingConflictError(false),
			xMsg:      "false",
			xConflict: false,
		},
		{
			name:      "customConflictTrue",
			err:       testingConflictError(true),
			xMsg:      "true",
			xConflict: true,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			assert.Check(t, cmp.Equal(IsConflict(c.err), c.xConflict))
			if c.err != nil {
				assert.Check(t, cmp.Equal(c.err.Error(), c.xMsg))
			}
		})
	}
}

func TestConflictCause(t *testing.T) {
	err := errors.New("test")
	e := &conflictError{err}
	assert.Check(t, cmp.Equal(e.Cause(), err))
	assert.Check(t, IsConflict(errors.Wrap(e, "some details")))
}
