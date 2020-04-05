package docker

import "github.com/cpuguy83/go-docker/system"

// SystemService creates a new system service from the client.
func (c *Client) SystemService() *system.Service {
	return system.NewService(c.tr)
}
