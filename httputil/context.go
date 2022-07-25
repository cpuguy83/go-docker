package httputil

import (
	"context"
	"io"
	"net/http"
)

var (
	// The default limit used by the client when reading response bodies.
	DefaultResponseLimit int64 = 16 * 1024
)

const (
	// UnlimitedResponseLimit is a value that can be used to indicate that a response should not be limited.
	UnlimitedResponseLimit int64 = -1
)

type responseLimit struct{}

// WithResponseLimit sets a limit for the max size to read from an http response.
// This value will be used by the client to limit how much data will be consumed from http responses.
func WithResponseLimit(ctx context.Context, limit int64) context.Context {
	return context.WithValue(ctx, responseLimit{}, limit)
}

// WithResponseLimitIfEmpty is like WithResponseLimit, but only sets a limit if none is set.
func WithResponseLimitIfEmpty(ctx context.Context, limit int64) context.Context {
	v := ctx.Value(responseLimit{})
	if v != nil {
		return ctx
	}
	return WithResponseLimit(ctx, limit)
}

// LimitResponse limits the size of the response body.
// This is used throughout the client to prevent a bad response from consuming too much memory.
// If a response limit is not set in the context, DefaultResponseLimit will be used.
//
// The value used is taken from the passed in context.
// Set this value by using:
//
// 	ctx = WithResponseLimit(ctx, limit)
func LimitResponse(ctx context.Context, resp *http.Response) {
	limit := DefaultResponseLimit
	v := ctx.Value(responseLimit{})
	if v != nil {
		limit = v.(int64)
	}
	if limit == UnlimitedResponseLimit {
		return
	}
	limited := io.LimitReader(resp.Body, limit)
	resp.Body = &wrapBody{limited, resp.Body}
}

type wrapBody struct {
	io.Reader
	io.Closer
}
