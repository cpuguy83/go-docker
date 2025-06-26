package transport

import (
	"context"
	"crypto/tls"
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

func TestTCPTransportDoRaw(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// for DoRaw test, we expect an upgrade request
		if req.Header.Get("Connection") == "Upgrade" {
			w.WriteHeader(http.StatusSwitchingProtocols)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	ctx := context.Background()

	u, err := url.Parse(srv.URL)
	assert.NilError(t, err)

	tr, err := TCPTransport(u.Host)
	assert.NilError(t, err)

	conn, err := tr.DoRaw(ctx, "POST", "/grpc", WithUpgrade("h2c"))
	assert.Assert(t, conn != nil, "expected a connection but got nil")
	assert.NilError(t, err)
}

func TestTCPTransportDoRawWithTLS(t *testing.T) {
	// test with TLS configuration
	tlsOpt := func(cfg *ConnectionConfig) error {
		cfg.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		return nil
	}

	tr, err := TCPTransport("localhost:2376", tlsOpt)
	assert.NilError(t, err)

	ctx := context.Background()

	_, err = tr.DoRaw(ctx, "POST", "/grpc", WithUpgrade("h2c"))
	// we expect this to fail with connection error since theres no server
	// but it should not panic
	assert.Assert(t, err != nil, "expected an error but got none")
}
