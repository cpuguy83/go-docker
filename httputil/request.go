package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/cpuguy83/go-docker/errdefs"
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
		return nil, errdefs.Wrap(err, "error doing request")
	}

	LimitResponse(ctx, resp)
	return resp, CheckResponseError(resp)
}

// WithJSONBody is a request option that sets the request body to the JSON encoded version of the passed in value.
func WithJSONBody(v interface{}) func(req *http.Request) error {
	return func(req *http.Request) error {
		data, err := json.Marshal(v)
		if err != nil {
			return errdefs.Wrap(err, "error marshaling json body")
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(data))
		if req.Header == nil {
			req.Header = http.Header{}
		}
		req.Header.Set("Content-Type", "application/json")
		return nil
	}
}
