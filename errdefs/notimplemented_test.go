package errdefs

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

type testingNotImplementedError bool

func (e testingNotImplementedError) Error() string {
	return fmt.Sprintf("%v", bool(e))
}

func (e testingNotImplementedError) NotImplemented() bool {
	return bool(e)
}

func TestIsNotImplemented(t *testing.T) {
	type testCase struct {
		name            string
		err             error
		xMsg            string
		xNotImplemented bool
	}

	for _, c := range []testCase{
		{
			name:            "NotImplementedf",
			err:             NotImplementedf("%s not found", "foo"),
			xMsg:            "foo not found",
			xNotImplemented: true,
		},
		{
			name:            "AsNotImplemented",
			err:             AsNotImplemented(errors.New("this is a test")),
			xMsg:            "this is a test",
			xNotImplemented: true,
		},
		{
			name:            "AsNotImplementedWithNil",
			err:             AsNotImplemented(nil),
			xMsg:            "",
			xNotImplemented: false,
		},
		{
			name:            "nilError",
			err:             nil,
			xMsg:            "",
			xNotImplemented: false,
		},
		{
			name:            "customNotImplementedFalse",
			err:             testingNotImplementedError(false),
			xMsg:            "false",
			xNotImplemented: false,
		},
		{
			name:            "customNotImplementedTrue",
			err:             testingNotImplementedError(true),
			xMsg:            "true",
			xNotImplemented: true,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			assert.Check(t, cmp.Equal(IsNotImplemented(c.err), c.xNotImplemented))
			if c.err != nil {
				assert.Check(t, cmp.Equal(c.err.Error(), c.xMsg))
			}
		})
	}
}

func TestNotImplementedCause(t *testing.T) {
	err := errors.New("test")
	e := &notImplementedError{err}
	assert.Check(t, cmp.Equal(e.Cause(), err))
	assert.Check(t, IsNotImplemented(errors.Wrap(e, "some details")))
}
