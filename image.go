package docker

import "github.com/cpuguy83/go-docker/image"

// ImageService provides access to image functionality, such as create, list.
func (c *Client) ImageService() *image.Service {
	return image.NewService(c.tr)
}
