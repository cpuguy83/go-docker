package container

import (
	"context"

	"github.com/cpuguy83/go-docker"
	"github.com/docker/docker/api/types"
)

func Inspect(ctx context.Context, name string) (types.ContainerJSON, error) {
	return docker.G(ctx).ContainerInspect(ctx, name)
}

func (c *container) Inspect(ctx context.Context) (types.ContainerJSON, error) {
	return Inspect(ctx, c.id)
}
