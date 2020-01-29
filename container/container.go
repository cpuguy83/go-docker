package container

import (
	"context"
	"io"

	"github.com/cpuguy83/go-docker"
	"github.com/docker/docker/api/types"
)

// Container is a wrapper for the conttainer client handlers.
// You can instantiate one by calling `Create` which will actually create a container in Docker or
// by calling `New` which only creates the object locally so you can interact with it in Docker.
type Container interface {
	Inspect(context.Context) (types.ContainerJSON, error)
	Logs(context.Context, ...LogsReadOption) (io.ReadCloser, error)
	Start(context.Context, ...StartOption) error
	Kill(context.Context, ...KillOption) error
	Stop(context.Context, ...StopOption) error
	ID() string
	StdinPipe(context.Context, ...AttachStdinOption) (io.WriteCloser, error)
	StdoutPipe(context.Context) (io.ReadCloser, error)
	StderrPipe(context.Context) (io.ReadCloser, error)
}

type container struct {
	id     string
	client *docker.Client
}

func (c *container) ID() string {
	return c.id
}
