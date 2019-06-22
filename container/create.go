package container

import (
	"context"

	"github.com/cpuguy83/go-docker"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

type CreateConfig struct {
	Config        *containertypes.Config
	HostConfig    *containertypes.HostConfig
	NetworkConfig *network.NetworkingConfig
	Name          string
}

type CreateOption func(*CreateConfig)

func WithCreateHostConfig(hc *containertypes.HostConfig) CreateOption {
	return func(c *CreateConfig) {
		c.HostConfig = hc
	}
}

func WithCreateConfig(cfg *containertypes.Config) CreateOption {
	return func(c *CreateConfig) {
		c.Config = cfg
	}
}

func WithCreateNetworkConfig(cfg *network.NetworkingConfig) CreateOption {
	return func(c *CreateConfig) {
		c.NetworkConfig = cfg
	}
}

func WithCreateName(name string) CreateOption {
	return func(c *CreateConfig) {
		c.Name = name
	}
}

func WithCreateImage(image string) CreateOption {
	return func(c *CreateConfig) {
		c.Config.Image = image
	}
}

func WithCreateCmd(cmd ...string) CreateOption {
	return func(c *CreateConfig) {
		c.Config.Cmd = cmd
	}
}

func WithCreateTTY(cfg *CreateConfig) {
	cfg.Config.Tty = true
}

func WithCreateAttachStdin(cfg *CreateConfig) {
	cfg.Config.AttachStdin = true
	cfg.Config.OpenStdin = true
}

func WithCreateStdinOnce(cfg *CreateConfig) {
	cfg.Config.StdinOnce = true
}

func WithCreateAttachStdout(cfg *CreateConfig) {
	cfg.Config.AttachStdout = true
}

func WithCreateAttachStderr(cfg *CreateConfig) {
	cfg.Config.AttachStderr = true
}

func Create(ctx context.Context, opts ...CreateOption) (Container, error) {
	c := CreateConfig{
		Config:        &containertypes.Config{},
		HostConfig:    &containertypes.HostConfig{},
		NetworkConfig: &network.NetworkingConfig{},
	}
	for _, o := range opts {
		o(&c)
	}

	resp, err := docker.G(ctx).ContainerCreate(ctx, c.Config, c.HostConfig, c.NetworkConfig, c.Name)
	if err != nil {
		return nil, err
	}

	return &container{id: resp.ID}, nil
}
