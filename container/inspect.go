package container

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/cpuguy83/go-docker"
	"github.com/docker/docker/api/types"
)

func Inspect(ctx context.Context, name string) (types.ContainerJSON, error) {
	var c types.ContainerJSON

	resp, err := docker.G(ctx).Do(ctx, http.MethodGet, "/containers/"+name+"/json")
	if err != nil {
		return c, err
	}

	defer resp.Body.Close()

	// TODO: Do not import from docker
	if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
		return c, errors.Wrap(err, "error decoding container inspect response")
	}
	return c, nil
}

func (c *container) Inspect(ctx context.Context) (types.ContainerJSON, error) {
	return Inspect(ctx, c.id)
}
