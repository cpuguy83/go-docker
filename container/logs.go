package container

import (
	"context"
	"io"

	"github.com/cpuguy83/go-docker"
	"github.com/docker/docker/api/types"
)

type ContainerLogsReadOption func(o *types.ContainerLogsOptions)

func (c *container) Logs(ctx context.Context, opts ...ContainerLogsReadOption) (io.ReadCloser, error) {
	return Logs(ctx, c.id, opts...)
}

// TODO: wrap the returned reader in a struct?
// TODO: Provide helper for consuming logs, maybe like daemon/logs does with a channel of disccrete log messages?
func Logs(ctx context.Context, name string, opts ...ContainerLogsReadOption) (io.ReadCloser, error) {
	var cfg types.ContainerLogsOptions
	for _, o := range opts {
		o(&cfg)
	}
	return docker.G(ctx).ContainerLogs(ctx, name, cfg)
}
