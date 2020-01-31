package docker

import (
	"github.com/cpuguy83/go-docker/container"
)

func (c *Client) ContainerService() *container.Service {
	return container.NewService(c.tr)
}
