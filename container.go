package docker

import (
	"github.com/cpuguy83/go-docker/container"
)

// ContainerService provides access to container functionaliaty, such as create, delete, start, stop, etc.
func (c *Client) ContainerService() *container.Service {
	return container.NewService(c.tr)
}
