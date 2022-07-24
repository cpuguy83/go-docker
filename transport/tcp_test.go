package transport

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"gotest.tools/v3/assert"
)

func TestTCPTransport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello " + req.URL.Path))
	}))
	defer srv.Close()

	ctx := context.Background()

	u, err := url.Parse(srv.URL)
	assert.NilError(t, err)

	tr, err := TCPTransport(u.Host)
	assert.NilError(t, err)

	resp, err := tr.Do(ctx, "GET", "/foo")
	assert.NilError(t, err)
	defer resp.Body.Close()

	data := "hello /foo"
	buf := make([]byte, len(data))
	_, err = io.ReadFull(resp.Body, buf)
	assert.NilError(t, err)
	assert.Equal(t, string(buf), data)
}
