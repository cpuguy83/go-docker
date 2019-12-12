package docker

import (
	"io"
	"net/http"

	"github.com/cpuguy83/go-docker/transport"
)

// WithRequestBody sets the body of the http request to the passed in reader
func WithRequestBody(r io.ReadCloser) transport.RequestOpt {
	return func(req *http.Request) error {
		req.Body = r
		return nil
	}
}
