package errdefs

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

type testingUnavailableError bool

func (e testingUnavailableError) Error() string {
	return fmt.Sprintf("%v", bool(e))
}

func (e testingUnavailableError) Unavailable() bool {
	return bool(e)
}

func TestIsUnavailable(t *testing.T) {
	type testCase struct {
		name         string
		err          error
		xMsg         string
		xUnavailable bool
	}

	for _, c := range []testCase{
		{
			name:         "Unavailablef",
			err:          Unavailablef("%s not found", "foo"),
			xMsg:         "foo not found",
			xUnavailable: true,
		},
		{
			name:         "AsUnavailable",
			err:          AsUnavailable(errors.New("this is a test")),
			xMsg:         "this is a test",
			xUnavailable: true,
		},
		{
			name:         "AsUnavailableWithNil",
			err:          AsUnavailable(nil),
			xMsg:         "",
			xUnavailable: false,
		},
		{
			name:         "nilError",
			err:          nil,
			xMsg:         "",
			xUnavailable: false,
		},
		{
			name:         "customUnavailableFalse",
			err:          testingUnavailableError(false),
			xMsg:         "false",
			xUnavailable: false,
		},
		{
			name:         "customUnavailableTrue",
			err:          testingUnavailableError(true),
			xMsg:         "true",
			xUnavailable: true,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			assert.Check(t, cmp.Equal(IsUnavailable(c.err), c.xUnavailable))
			if c.err != nil {
				assert.Check(t, cmp.Equal(c.err.Error(), c.xMsg))
			}
		})
	}
}

func TestUnavailableCause(t *testing.T) {
	err := errors.New("test")
	e := &unavailableError{err}
	assert.Check(t, cmp.Equal(e.Cause(), err))
	assert.Check(t, IsUnavailable(errors.Wrap(e, "some details")))
}
