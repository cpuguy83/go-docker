package transport

import (
	"context"
	"net"
	"net/http"
)

func TCPTransport(host string, opts ...ConnectionOption) (*Transport, error) {
	var cfg ConnectionConfig

	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return nil, err
		}
	}

	httpTransport := &http.Transport{
		DialContext:     new(net.Dialer).DialContext,
		TLSClientConfig: cfg.TLSConfig,
	}

	scheme := "http"
	if cfg.TLSConfig != nil {
		scheme = "https"
	}

	dial := func(ctx context.Context) (net.Conn, error) {
		return httpTransport.DialContext(ctx, "tcp", host)
	}

	t := &Transport{
		scheme: scheme,
		host:   host,
		c: &http.Client{
			Transport: httpTransport,
		},
		dial: dial,
	}

	return t, nil
}
