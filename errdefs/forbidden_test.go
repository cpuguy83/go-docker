package errdefs

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

type testingForbiddenError bool

func (e testingForbiddenError) Error() string {
	return fmt.Sprintf("%v", bool(e))
}

func (e testingForbiddenError) Forbidden() bool {
	return bool(e)
}

func TestIsForbidden(t *testing.T) {
	type testCase struct {
		name       string
		err        error
		xMsg       string
		xForbidden bool
	}

	for _, c := range []testCase{
		{
			name:       "Forbiddenf",
			err:        Forbiddenf("%s not found", "foo"),
			xMsg:       "foo not found",
			xForbidden: true,
		},
		{
			name:       "AsForbidden",
			err:        AsForbidden(errors.New("this is a test")),
			xMsg:       "this is a test",
			xForbidden: true,
		},
		{
			name:       "AsForbiddenWithNil",
			err:        AsForbidden(nil),
			xMsg:       "",
			xForbidden: false,
		},
		{
			name:       "nilError",
			err:        nil,
			xMsg:       "",
			xForbidden: false,
		},
		{
			name:       "customForbiddenFalse",
			err:        testingForbiddenError(false),
			xMsg:       "false",
			xForbidden: false,
		},
		{
			name:       "customForbiddenTrue",
			err:        testingForbiddenError(true),
			xMsg:       "true",
			xForbidden: true,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			assert.Check(t, cmp.Equal(IsForbidden(c.err), c.xForbidden))
			if c.err != nil {
				assert.Check(t, cmp.Equal(c.err.Error(), c.xMsg))
			}
		})
	}
}

func TestForbiddenCause(t *testing.T) {
	err := errors.New("test")
	e := &forbiddenError{err}
	assert.Check(t, cmp.Equal(e.Cause(), err))
	assert.Check(t, IsForbidden(errors.Wrap(e, "some details")))
}
