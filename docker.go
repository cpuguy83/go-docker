package docker

import (
	"context"

	"github.com/cpuguy83/go-docker/transport"
)

type clientKey struct{}

func WithClient(ctx context.Context, c *Client) context.Context {
	return context.WithValue(ctx, clientKey{}, c)
}

func G(ctx context.Context) *Client {
	return GetClient(ctx)
}

func GetClient(ctx context.Context) *Client {
	if c := ClientFromContext(ctx); c != nil {
		return c
	}
	t := transport.DefaultUnixTransport()
	return &Client{
		Transport: t,
	}
}

func ClientFromContext(ctx context.Context) *Client {
	c := ctx.Value(clientKey{})
	if c != nil {
		return c.(*Client)
	}
	return nil
}
