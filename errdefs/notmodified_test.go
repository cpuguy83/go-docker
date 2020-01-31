package errdefs

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

type testingNotModifiedError bool

func (e testingNotModifiedError) Error() string {
	return fmt.Sprintf("%v", bool(e))
}

func (e testingNotModifiedError) NotModified() bool {
	return bool(e)
}

func TestIsNotModified(t *testing.T) {
	type testCase struct {
		name         string
		err          error
		xMsg         string
		xNotModified bool
	}

	for _, c := range []testCase{
		{
			name:         "NotModifiedf",
			err:          NotModifiedf("%s not found", "foo"),
			xMsg:         "foo not found",
			xNotModified: true,
		},
		{
			name:         "AsNotModified",
			err:          AsNotModified(errors.New("this is a test")),
			xMsg:         "this is a test",
			xNotModified: true,
		},
		{
			name:         "AsNotModifiedWithNil",
			err:          AsNotModified(nil),
			xMsg:         "",
			xNotModified: false,
		},
		{
			name:         "nilError",
			err:          nil,
			xMsg:         "",
			xNotModified: false,
		},
		{
			name:         "customNotModifiedFalse",
			err:          testingNotModifiedError(false),
			xMsg:         "false",
			xNotModified: false,
		},
		{
			name:         "customNotModifiedTrue",
			err:          testingNotModifiedError(true),
			xMsg:         "true",
			xNotModified: true,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			assert.Check(t, cmp.Equal(IsNotModified(c.err), c.xNotModified))
			if c.err != nil {
				assert.Check(t, cmp.Equal(c.err.Error(), c.xMsg))
			}
		})
	}
}

func TestNotModifiedCause(t *testing.T) {
	err := errors.New("test")
	e := &notModifiedError{err}
	assert.Check(t, cmp.Equal(e.Cause(), err))
	assert.Check(t, IsNotModified(errors.Wrap(e, "some details")))
}
