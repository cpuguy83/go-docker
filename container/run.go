package container

import (
	"context"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

type RunConfig struct {
	CreateConfig CreateConfig
	StartConfig  types.ContainerStartOptions
}

type RunOption func(*RunConfig)

func WithRunCreateOption(co CreateOption) RunOption {
	return func(c *RunConfig) {
		co(&(c.CreateConfig))
	}
}

func WithRunStartOption(co StartOption) RunOption {
	return func(c *RunConfig) {
		co(&(c.StartConfig))
	}
}

func Run(ctx context.Context, opts ...RunOption) (Container, error) {
	cfg := RunConfig{
		CreateConfig: CreateConfig{
			Config:        &containertypes.Config{},
			HostConfig:    &containertypes.HostConfig{},
			NetworkConfig: &network.NetworkingConfig{},
		},
	}
	for _, o := range opts {
		o(&cfg)
	}

	c, err := Create(ctx,
		WithCreateConfig(cfg.CreateConfig.Config),
		WithCreateHostConfig(cfg.CreateConfig.HostConfig),
		WithCreateNetworkConfig(cfg.CreateConfig.NetworkConfig),
	)
	if err != nil {
		return nil, err
	}

	return c, c.Start(ctx, func(o *types.ContainerStartOptions) {
		*o = cfg.StartConfig
	})
}
