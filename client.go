package docker

import (
	"github.com/cpuguy83/go-docker/transport"
)

// Client is the main docker client
// Create one with `NewClient`
type Client struct {
	tr transport.Doer
}

// NewClientConfig is the list of options for configuring a new docker client
type NewClientConfig struct {
	// Transport is the communication method for reaching a docker engine instance.
	// You can implement your own transport, or use the ones provided in the transport package.
	// If this is unset, the default transport will be used (unix socket connected to /var/run/docker.sock).
	Transport transport.Doer
}

type NewClientOption func(*NewClientConfig)

// NewClient creates a new docker client
// You can pass in options using functional arguments.
//
// If no transport is provided as an option, the default transport will be used.
//
// You probably want to set an API version for the client to use here.
// See `NewClientConfig` for available options
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
	return &Client{tr: tr}
}

// WithTransport is a NewClientOption that sets the transport to be used for the client.
func WithTransport(tr transport.Doer) NewClientOption {
	return func(cfg *NewClientConfig) {
		cfg.Transport = tr
	}
}
