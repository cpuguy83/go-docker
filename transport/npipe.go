package transport

import (
	"context"
	"net"
	"net/http"
	"net/url"
)

func DefaultWindowsTransport() *Transport {
	t, _ := NpipeTransport("//./pipe/docker_engine")
	return t
}

func NpipeTransport(path string, opts ...ConnectionOption) (*Transport, error) {
	var cfg ConnectionConfig

	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return nil, err
		}
	}

	t := &http.Transport{
		DisableCompression: true,
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return winDailer(path, nil)
		},
	}

	dail := func(ctx context.Context) (net.Conn, error) {
		return t.DialContext(ctx, "", "")
	}

	return &Transport{
		host:   url.PathEscape(path),
		scheme: "http",
		c: &http.Client{
			Transport: t,
		},
		dial: dail,
	}, nil
}
