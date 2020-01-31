package container

import (
	"context"

	"github.com/cpuguy83/go-docker/transport"
)

// Container provides bindings for interacting with a container in Docker
type Container struct {
	id string
	tr transport.Doer
}

// ID returns the container ID
func (c *Container) ID() string {
	return c.id
}

// NewConfig holds the options available for `New`
type NewConfig struct {
}

// NewOption is used as functional parameters to `New`
type NewOption func(*NewConfig)

// New creates a new container object in memory. This function does not interact with the Docker API at all.
// If the container does not exist in Docker, all calls on the Container will fail.
//
// To actually create a container you must call `Create` first (which will return a container object to you).
//
// TODO: This may well change to actually inspect the container and fetch the actual container ID.
func (s *Service) NewContainer(_ context.Context, id string, opts ...NewOption) *Container {
	var cfg NewConfig
	for _, o := range opts {
		o(&cfg)
	}
	return &Container{id: id, tr: s.tr}
}
