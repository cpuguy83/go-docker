package container

import (
	"context"
	"time"

	"github.com/cpuguy83/go-docker"
)

type StopOption func(*StopConfig)

type StopConfig struct {
	Timeout *time.Duration
}

func Stop(ctx context.Context, name string, opts ...StopOption) error {
	var cfg StopConfig
	for _, o := range opts {
		o(&cfg)
	}
	return docker.G(ctx).ContainerStop(ctx, name, cfg.Timeout)
}

func (c *container) Stop(ctx context.Context, opts ...StopOption) error {
	return Stop(ctx, c.id, opts...)
}
