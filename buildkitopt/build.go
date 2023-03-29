/*
Package buildkit provides the neccessary functionality to create a buildkit
client that can be used to communicate with the buildkit service provided by
dockerd.

This is provided in a module separate from the main go-docker module so that
only those that need it will pull in the buildkit dependencies.
*/

package buildkitopt

import (
	"context"
	"net"
	"net/http"

	"github.com/cpuguy83/go-docker/transport"
	"github.com/cpuguy83/go-docker/version"
	"github.com/moby/buildkit/client"
)

// NewClient is a convenience wrapper which creates a buildkit client for the
// buildkit service provided by dockerd. This just wraps buildkit's client.New
// to include WithSessionDialer and WithGRPCDialer automatically in addition to
// the opts provided.
func NewClient(ctx context.Context, tr transport.Doer, opts ...client.ClientOpt) (*client.Client, error) {
	return client.New(ctx, "", append(opts, FromDocker(tr)...)...)
}

// FromDocker is a convenience function that returns a slice of ClientOpts that can be used to create a client for the buildkit GRPC and session APIs provided by dockerd.
func FromDocker(tr transport.Doer) []client.ClientOpt {
	return []client.ClientOpt{
		WithGRPCDialer(tr),
		WithSessionDialer(tr),
	}
}

// WithSessionDialer creates a ClientOpt that can be used to create a client for the buildkit session API provided by dockerd.
func WithSessionDialer(tr transport.Doer) client.ClientOpt {
	return client.WithSessionDialer(func(ctx context.Context, proto string, meta map[string][]string) (net.Conn, error) {
		return tr.DoRaw(ctx, http.MethodPost, version.Join(ctx, "/session"), transport.WithUpgrade(proto), transport.WithAddHeaders(meta))
	})
}

// WithGRPCDialer creates a ClientOpt that can be used to create a client for the buildkit GRPC API provided by dockerd.
func WithGRPCDialer(tr transport.Doer) client.ClientOpt {
	return client.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
		return tr.DoRaw(ctx, http.MethodPost, version.Join(ctx, "/grpc"), transport.WithUpgrade("h2c"))
	})
}
