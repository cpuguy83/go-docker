package transport

import (
	"context"
	"io"
	"net/http"
)

// VersionedTransport wraps a Doer with a new Doer that injects the specified API version
type Versioned struct {
	Transport  Doer
	APIVersion string
}

func (t *Versioned) Do(ctx context.Context, method, uri string, opts ...RequestOpt) (*http.Response, error) {
	if t.APIVersion != "" {
		uri = "/v" + t.APIVersion + uri
	}
	return t.Transport.Do(ctx, method, uri, opts...)
}

func (t *Versioned) DoRaw(ctx context.Context, method, uri string, opts ...RequestOpt) (io.ReadWriteCloser, error) {
	if t.APIVersion != "" {
		uri = "/v" + t.APIVersion + uri
	}
	return t.Transport.DoRaw(ctx, method, uri, opts...)
}
