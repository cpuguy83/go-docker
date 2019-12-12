package system

import (
	"context"

	"github.com/cpuguy83/go-docker"
	"github.com/docker/docker/api/types"
)

func Ping(ctx context.Context) (types.Ping, error) {
	return docker.G(ctx).Ping(ctx)
}
