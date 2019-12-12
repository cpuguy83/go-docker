package transport

import (
	"context"
	"net"
	"net/http"
)

func DefaultUnixTransport() *Transport {
	t, _ := UnixSocketTransport("/var/run/docker.sock")
	return t
}

// UnixSocketTransport creates a Transport that works for unix sockets.
//
// Note: This will attempt to use the TLSConfig if it is set on the connection options
// If you do not want to use TLS, do not set it on the connection options.
func UnixSocketTransport(sock string, opts ...ConnectionOption) (*Transport, error) {
	var cfg ConnectionConfig
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return nil, err
		}
	}

	t := &http.Transport{
		DisableCompression: true,
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return new(net.Dialer).DialContext(ctx, "unix", sock)
		},
		TLSClientConfig: cfg.TLSConfig,
	}

	scheme := "http"
	if cfg.TLSConfig != nil {
		scheme = "https"
	}

	dial := func(ctx context.Context) (net.Conn, error) {
		return t.DialContext(ctx, "", "")
	}

	return &Transport{
		host:   sock,
		scheme: scheme,
		c: &http.Client{
			Transport: t,
		},
		dial: dial,
	}, nil
}
