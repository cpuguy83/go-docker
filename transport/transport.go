package transport

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"

	"github.com/pkg/errors"
)

// Doer performs an http request for Client
// It is the Doer's responsibility to deal with setting the host details on
// the request
// It is expected that one Doer connects to one Docker instance.
type Doer interface {
	// Do typically performs a normal http request/response
	Do(ctx context.Context, method string, uri string, opts ...RequestOpt) (*http.Response, error)
	// DoRaw performs the request but passes along the response as a bi-directional stream
	DoRaw(ctx context.Context, method string, uri string, opts ...RequestOpt) (io.ReadWriteCloser, error)
}

// WithRequestBody sets the body of the http request to the passed in reader
func WithRequestBody(r io.ReadCloser) RequestOpt {
	return func(req *http.Request) error {
		req.Body = r
		return nil
	}
}

// Transport implements the Doer interface for all the normal docker protocols).
// This would normally be things that would go over a net.Conn, such as unix or tcp sockets.
//
// Create a transport from one of the available helper functions.
type Transport struct {
	c      *http.Client
	dial   func(context.Context) (net.Conn, error)
	host   string
	scheme string
}

// RequestOpt is as functional arguments to configure an HTTP request for a Doer.
type RequestOpt func(*http.Request) error

// Do implements the Doer.Do interface
func (t *Transport) Do(ctx context.Context, method, uri string, opts ...RequestOpt) (*http.Response, error) {
	req := &http.Request{}
	req.Method = method
	req.URL = &url.URL{Path: uri, Host: t.host, Scheme: t.scheme}

	req = req.WithContext(ctx)

	for _, o := range opts {
		if err := o(req); err != nil {
			return nil, err
		}
	}
	resp, err := t.c.Do(req)
	if err != nil {
		return resp, err
	}
	if err := checkResponseError(resp); err != nil {
		resp.Body.Close()
		return nil, err
	}
	return resp, nil
}

// Do implements the Doer.DoRaw interface
func (t *Transport) DoRaw(ctx context.Context, method, uri string, opts ...RequestOpt) (rwc io.ReadWriteCloser, retErr error) {
	req := &http.Request{Header: http.Header{}}
	req.Method = method
	req.URL = &url.URL{Path: uri, Host: t.host, Scheme: t.scheme}
	req.Header.Set("Connection", "Upgrade")
	proto := "tcp" // # TODO: This is not right but it's what the official docker client currently does.
	req.Header.Set("Upgrade", proto)

	req = req.WithContext(ctx)

	for _, o := range opts {
		if err := o(req); err != nil {
			return nil, err
		}
	}

	conn, err := t.dial(ctx)
	if err != nil {
		return nil, err
	}

	cc := httputil.NewClientConn(conn, nil)
	if retErr != nil {
		cc.Close()
	}

	resp, err := cc.Do(req)
	if err != httputil.ErrPersistEOF {
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusSwitchingProtocols {
			resp.Body.Close()
			return nil, errors.Errorf("unable to upgrade to %s, received %d", proto, resp.StatusCode)
		}
	}

	conn, buf := cc.Hijack()
	return newHijackedConn(conn, buf), nil
}

type closeWriter interface {
	CloseWrite() error
}

// FromConnectionString creates a transport from the provided connection string
// This connection string is the one defined in the official docker client for DOCKER_HOST
func FromConnectionString(s string, opts ...ConnectionOption) (*Transport, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	return FromConnectionURL(u, opts...)
}

// ConnectionOption is use as functional arguments for creating a Transport
// It configures a ConnectionConfig
type ConnectionOption func(*ConnectionConfig) error

// ConnectionConfig holds the options available for configuring a new transport.
type ConnectionConfig struct {
	TLSConfig *tls.Config
}

// FromConnectionURL creates a Transport from a provided URL
//
// The URL's scheme must specify the protocol ("unix", "tcp", etc.)
//
// TODO: implement named pipes (windows) and ssh schemes.
func FromConnectionURL(u *url.URL, opts ...ConnectionOption) (*Transport, error) {
	switch u.Scheme {
	case "unix":
		return UnixSocketTransport(path.Join(u.Host, u.Path), opts...)
	case "tcp":
		return TCPTransport(u.Host, opts...)
	default:
		// TODO: npipe, ssh
		return nil, errors.Errorf("protocol not supported: %s", u.Scheme)
	}
}
