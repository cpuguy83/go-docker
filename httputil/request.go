package httputil

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

// DoRequest performs the passed in function, passing it the provided context.
// The response body reader is then limited and checked for error status codes.
// The returned response body will always be limited by the limit value set in the passed in context.
//
// In the case that an error is found, the response body will be closed.
// This may return a non-nil response even on error.
// This allows callers to inspect response headers and other things.
func DoRequest(ctx context.Context, do func(context.Context) (*http.Response, error)) (*http.Response, error) {
	resp, err := do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error doing request")
	}

	LimitResponse(ctx, resp)
	return resp, CheckResponseError(resp)
}
