package errdefs

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

type testingUnauthorizedError bool

func (e testingUnauthorizedError) Error() string {
	return fmt.Sprintf("%v", bool(e))
}

func (e testingUnauthorizedError) Unauthorized() bool {
	return bool(e)
}

func TestIsUnauthorized(t *testing.T) {
	type testCase struct {
		name          string
		err           error
		xMsg          string
		xUnauthorized bool
	}

	for _, c := range []testCase{
		{
			name:          "Unauthorizedf",
			err:           Unauthorizedf("%s not found", "foo"),
			xMsg:          "foo not found",
			xUnauthorized: true,
		},
		{
			name:          "AsUnauthorized",
			err:           AsUnauthorized(errors.New("this is a test")),
			xMsg:          "this is a test",
			xUnauthorized: true,
		},
		{
			name:          "AsUnauthorizedWithNil",
			err:           AsUnauthorized(nil),
			xMsg:          "",
			xUnauthorized: false,
		},
		{
			name:          "nilError",
			err:           nil,
			xMsg:          "",
			xUnauthorized: false,
		},
		{
			name:          "customUnauthorizedFalse",
			err:           testingUnauthorizedError(false),
			xMsg:          "false",
			xUnauthorized: false,
		},
		{
			name:          "customUnauthorizedTrue",
			err:           testingUnauthorizedError(true),
			xMsg:          "true",
			xUnauthorized: true,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			assert.Check(t, cmp.Equal(IsUnauthorized(c.err), c.xUnauthorized))
			if c.err != nil {
				assert.Check(t, cmp.Equal(c.err.Error(), c.xMsg))
			}
		})
	}
}

func TestUnauthorizedCause(t *testing.T) {
	err := errors.New("test")
	e := &unauthorizedError{err}
	assert.Check(t, cmp.Equal(e.Cause(), err))
	assert.Check(t, IsUnauthorized(errors.Wrap(e, "some details")))
}
