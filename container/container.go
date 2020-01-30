package container

import (
	"context"

	"github.com/cpuguy83/go-docker/transport"
)

type Container struct {
	id string
	tr transport.Doer
}

func (c *Container) ID() string {
	return c.id
}

// NewConfig holds the options available for `New`
type NewConfig struct {
}

// NewOption is used as functional parameters to `New`
type NewOption func(*NewConfig)

// NewContainer instantiates a container object that you can interact with the Service's transport
// See the unbounded `New` function for more details
func (s *Service) NewContainer(ctx context.Context, name string, opts ...NewOption) *Container {
	return New(ctx, s.tr, name, opts...)
}

// New creates a new container object in memory, it does not interact with the Docker API at all.
// If the container does not exist in Docker, all calls on the Container will fail.
//
// To actually create a container you must call `Create` on the container service.
func New(_ context.Context, tr transport.Doer, name string, opts ...NewOption) *Container {
	var cfg NewConfig
	for _, o := range opts {
		o(&cfg)
	}
	return &Container{id: name, tr: tr}
}
