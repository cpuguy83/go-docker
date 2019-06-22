package container

import (
	"context"

	"github.com/cpuguy83/go-docker"
	"github.com/docker/docker/api/types"
)

type StartOption func(*types.ContainerStartOptions)

func (c *container) Start(ctx context.Context, opts ...StartOption) error {
	return Start(ctx, c.id, opts...)
}

func Start(ctx context.Context, name string, opts ...StartOption) error {
	var cfg types.ContainerStartOptions
	for _, o := range opts {
		o(&cfg)
	}
	return docker.G(ctx).ContainerStart(ctx, name, cfg)
}
