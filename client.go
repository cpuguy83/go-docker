package docker

import (
	"github.com/cpuguy83/go-docker/transport"
)

type Client struct {
	tr transport.Doer
}

type NewClientConfig struct {
	APIVersion string
	Transport  transport.Doer
}

type NewClientOption func(*NewClientConfig)

func NewClient(opts ...NewClientOption) *Client {
	var cfg NewClientConfig
	for _, o := range opts {
		o(&cfg)
	}
	tr := cfg.Transport
	if tr == nil {
		// TODO: make this platform specific
		tr = transport.DefaultUnixTransport()
	}
	if cfg.APIVersion != "" {
		return &Client{tr: &transport.Versioned{APIVersion: cfg.APIVersion, Transport: tr}}
	}
	return &Client{tr: tr}
}


