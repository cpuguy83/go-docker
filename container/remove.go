package container

import (
	"context"

	"github.com/cpuguy83/go-docker"
	"github.com/docker/docker/api/types"
)

type RemoveOption func(*types.ContainerRemoveOptions)

func WithRemoveForce(o *types.ContainerRemoveOptions) {
	o.Force = true
}

func Remove(ctx context.Context, name string, opts ...RemoveOption) error {
	var cfg types.ContainerRemoveOptions
	for _, o := range opts {
		o(&cfg)
	}
	return docker.G(ctx).ContainerRemove(ctx, name, cfg)
}
