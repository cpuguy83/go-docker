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

// NewConfig holds the options available for `New`
type NewConfig struct {
	Client *docker.Client
}

// NewOption is used as functional parameters to `New`
type NewOption func(*NewConfig)

// New instantiates a container object that you can interact with via the provided client.
// New does not create any resources in docker or even hit the API.
func New(ctx context.Context, name string, opts ...NewOption) (Container, error) {
	var cfg NewConfig
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.Client == nil {
		cfg.Client = docker.G(ctx)
	}
	c := &container{id: name, client: cfg.Client}
	return c, nil
}
