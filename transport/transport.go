package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"time"
)

// Doer performs an http request for Client
// It is the Doer's responsibility to deal with setting the host details on
// the request
// It is expected that one Doer connects to one Docker instance.
type Doer interface {
	// Do typically performs a normal http request/response
	Do(ctx context.Context, method string, uri string, opts ...RequestOpt) (*http.Response, error)
	// DoRaw performs the request but passes along the response as a bi-directional stream
	DoRaw(ctx context.Context, method string, uri string, opts ...RequestOpt) (net.Conn, error)
}

// RequestOpt is as functional arguments to configure an HTTP request for a Doer.
type RequestOpt func(*http.Request) error

// Transport implements the Doer interface for all the normal docker protocols).
// This would normally be things that would go over a net.Conn, such as unix or tcp sockets.
//
// Create a transport from one of the available helper functions.
type Transport struct {
	c         *http.Client
	dial      func(context.Context) (net.Conn, error)
	host      string
	scheme    string
	transform func(*http.Request)
}

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

	if t.transform != nil {
		t.transform(req)
	}
	resp, err := t.c.Do(req)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// Do implements the Doer.DoRaw interface
func (t *Transport) DoRaw(ctx context.Context, method, uri string, opts ...RequestOpt) (conn net.Conn, retErr error) {
	req := &http.Request{Header: http.Header{}}
	req.Method = method
	req.URL = &url.URL{Path: uri, Host: t.host, Scheme: t.scheme}

	req = req.WithContext(ctx)

	for _, o := range opts {
		if err := o(req); err != nil {
			return nil, err
		}
	}

	if t.transform != nil {
		t.transform(req)
	}

	conn, err := t.dial(ctx)
	if err != nil {
		return nil, err
	}

	// There can be long periods of inactivity when hijacking a connection.
	// Set keep-alive to ensure that the connection is not broken due to idle time.
	if tc, ok := conn.(*net.TCPConn); ok {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(30 * time.Second)
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
			return nil, fmt.Errorf("unable to upgrade to %s, received %d", req.Header.Get("Upgrade"), resp.StatusCode)
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
// TODO: implement ssh schemes.
func FromConnectionURL(u *url.URL, opts ...ConnectionOption) (*Transport, error) {
	switch u.Scheme {
	case "unix":
		return UnixSocketTransport(path.Join(u.Host, u.Path), opts...)
	case "tcp":
		return TCPTransport(u.Host, opts...)
	case "npipe":
		return NpipeTransport(u.Path, opts...)
	default:
		// TODO: ssh
		return nil, fmt.Errorf("protocol not supported: %s", u.Scheme)
	}
}

const (
	headerConnection = "Connection"
	headerUpgrade    = "Upgrade"
)

// WithUpgrade is a RequestOpt that sets the request to upgrade to the specified protocol.
func WithUpgrade(proto string) RequestOpt {
	return func(req *http.Request) error {
		req.Header.Set(headerConnection, headerUpgrade)
		req.Header.Set(headerUpgrade, proto)
		return nil
	}
}

// WithAddHeaders is a RequestOpt that adds the specified headers to the request.
// If the header already exists, it will be appended to.
func WithAddHeaders(headers map[string][]string) RequestOpt {
	return func(req *http.Request) error {
		for k, v := range headers {
			for _, vv := range v {
				req.Header.Add(k, vv)
			}
		}
		return nil
	}
}

// go1.20.6 introduced a breaking change which makes paths an invalid value for a host header
// This is problematic for us because we use the path as the URI for the request.
// If req.Host is not set OR is the same as the socket path (basically unmodified by something else) then we can rewrite it.
// If its anything else then this was changed by something else and we should not touch it.
func go120Dot6HostTransform(sock string) func(req *http.Request) {
	return func(req *http.Request) {
		if req.Host == "" || req.Host == sock {
			req.Host = strings.Replace(sock, "/", "_", -1)
		}
	}
}
