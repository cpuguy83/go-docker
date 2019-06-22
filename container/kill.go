package container

import (
	"context"

	"github.com/cpuguy83/go-docker"
)

type KillOption func(*KillConfig)

type KillConfig struct {
	Signal string
}

func Kill(ctx context.Context, name string, opts ...KillOption) error {
	var cfg KillConfig
	for _, o := range opts {
		o(&cfg)
	}
	return docker.G(ctx).ContainerKill(ctx, name, cfg.Signal)
}

func (c *container) Kill(ctx context.Context, opts ...KillOption) error {
	return Kill(ctx, c.id, opts...)
}
