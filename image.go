package docker

import "github.com/cpuguy83/go-docker/image"

// ImageService provides access to image functionaliaty, such as create, list.
func (c *Client) ImageService() *image.Service {
	return image.NewService(c.tr)
}
