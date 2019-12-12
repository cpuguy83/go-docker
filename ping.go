package docker

import (
	"context"

	"github.com/docker/docker/api/types"
)

func (c *Client) Ping(ctx context.Context) (types.Ping, error) {
	getAPIVersion(ctx, c.APIVersion)

	resp, err := c.Do(ctx, "GET", "/_ping")
	if err != nil {
		return types.Ping{}, err
	}
	defer resp.Body.Close()

	var p types.Ping
	p.APIVersion = resp.Header.Get("API-Version")
	p.OSType = resp.Header.Get("OSType")
	p.Experimental = resp.Header.Get("Docker-Experimental") == "true"
	p.BuilderVersion = types.BuilderVersion(resp.Header.Get("Builder-Version"))

	return p, checkResponseError(resp)
}
