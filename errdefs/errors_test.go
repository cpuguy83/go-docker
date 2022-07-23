package errdefs

import (
	"errors"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type testCase struct {
	Err  error
	New  func(string) error
	Newf func(string, ...interface{}) error
	Is   func(error) bool
	As   func(error) error
}

func TestErrors(t *testing.T) {
	cases := map[string]testCase{
		"conflict":        {ErrConflict, Conflict, Conflictf, IsConflict, AsConflict},
		"forbidden":       {ErrForbidden, Forbidden, Forbiddenf, IsForbidden, AsForbidden},
		"invalid":         {ErrInvalid, Invalid, Invalidf, IsInvalid, AsInvalid},
		"not found":       {ErrNotFound, NotFound, NotFoundf, IsNotFound, AsNotFound},
		"not implemented": {ErrNotImplemented, NotImplemented, NotImplementedf, IsNotImplemented, AsNotImplemented},
		"not modified":    {ErrNotModified, NotModified, NotModifiedf, IsNotModified, AsNotModified},
		"unauthorized":    {ErrUnauthorized, Unauthorized, Unauthorizedf, IsUnauthorized, AsUnauthorized},
		"unavailable":     {ErrUnavailable, Unavailable, Unavailablef, IsUnavailable, AsUnavailable},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, testError(tc))
	}
}

func getFunctionName(f interface{}) string {
	n := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	return n[strings.LastIndexAny(n, ".")+1:]
}

func testError(tc testCase) func(t *testing.T) {
	return func(t *testing.T) {
		e := tc.New("name")
		if !errors.Is(e, tc.Err) {
			t.Fatalf("%s: expected %v to be %v", getFunctionName(tc.New), e, tc.Err)
		}
		if !tc.Is(e) {
			t.Fatalf("%s did not return true after creating error with %s", getFunctionName(tc.Is), getFunctionName(tc.New))
		}

		e = tc.Newf("name %s", "value")
		if !errors.Is(e, tc.Err) {
			t.Fatalf("%s: expected %v to be %v", getFunctionName(tc.Newf), e, tc.Err)
		}
		if !tc.Is(e) {
			t.Fatalf("%s did not return true after creating error with %s", getFunctionName(tc.Is), getFunctionName(tc.Newf))
		}

		e = errors.New(t.Name())
		e = tc.As(e)
		if !errors.Is(e, tc.Err) {
			t.Fatalf("expected error to be wrapped by %v", tc.Err)
		}
		if !tc.Is(e) {
			t.Fatalf("%s did not return true after wrapping error with %s", getFunctionName(tc.Is), getFunctionName(tc.As))
		}
	}
}
