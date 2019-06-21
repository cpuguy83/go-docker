package docker

import (
	"context"

	"github.com/docker/docker/client"
)

type clientKey struct{}

func WithClient(ctx context.Context, c *client.Client) context.Context {
	return context.WithValue(ctx, clientKey{}, c)
}

func G(ctx context.Context) *client.Client {
	return GetClient(ctx)
}

func GetClient(ctx context.Context) *client.Client {
	if c := ClientFromContext(ctx); c != nil {
		return c
	}
	if c, _ := client.NewEnvClient(); c != nil {
		return c
	}
	panic("nil client")
}

func ClientFromContext(ctx context.Context) *client.Client {
	c := ctx.Value(clientKey{})
	if c != nil {
		return c.(*client.Client)
	}
	return nil
}
