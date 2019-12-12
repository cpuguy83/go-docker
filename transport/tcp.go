package transport

import (
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

	t := &Transport{
		scheme: "http",
		host:   host,
		c: &http.Client{
			Transport: &http.Transport{
				DialContext:     new(net.Dialer).DialContext,
				TLSClientConfig: cfg.TLSConfig,
			},
		},
	}

	if cfg.TLSConfig != nil {
		t.scheme = "https"
	}

	return t, nil
}
