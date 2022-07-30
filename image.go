package docker

import "github.com/cpuguy83/go-docker/image"

// SystemService creates a new system service from the client.
func (c *Client) ImageService() *image.Service {
	return image.NewService(c.tr)
}
