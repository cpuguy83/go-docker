package container

import (
	"context"

	"github.com/cpuguy83/go-docker"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

type CreateRequest struct {
	Config           *containertypes.Config
	HostConfig       *containertypes.HostConfig
	NetworkingConfig *network.NetworkingConfig
	Name             string
}

type CreateOpt func(*CreateRequest)

func WithCreateHostConfig(hc *containertypes.HostConfig) CreateOpt {
	return func(c *CreateRequest) {
		c.HostConfig = hc
	}
}

func WithCreateConfig(cfg *containertypes.Config) CreateOpt {
	return func(c *CreateRequest) {
		c.Config = cfg
	}
}

func WithCreateName(name string) CreateOpt {
	return func(c *CreateRequest) {
		c.Name = name
	}
}

func WithCreateImage(image string) CreateOpt {
	return func(c *CreateRequest) {
		c.Config.Image = image
	}
}

func Create(ctx context.Context, opts ...CreateOpt) (Container, error) {
	c := CreateRequest{
		Config:           &containertypes.Config{},
		HostConfig:       &containertypes.HostConfig{},
		NetworkingConfig: &network.NetworkingConfig{},
	}
	for _, o := range opts {
		o(&c)
	}

	resp, err := docker.G(ctx).ContainerCreate(ctx, c.Config, c.HostConfig, c.NetworkingConfig, c.Name)
	if err != nil {
		return nil, err
	}

	return &container{id: resp.ID}, nil
}
