package container

import (
	"context"
	"io"
)

type Container interface {
	Logs(context.Context, ...ContainerLogsReadOption) (io.ReadCloser, error)
	Start(context.Context, ...ContainerStartOption) error
	ID() string
}

type container struct {
	id string
}

func (c *container) ID() string {
	return c.id
}
