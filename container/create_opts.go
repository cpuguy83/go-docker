package container

import (
	"github.com/cpuguy83/go-docker/container/containerapi"
)

type CreateOption func(*CreateConfig)

func WithCreateHostConfig(hc *containerapi.HostConfig) CreateOption {
	return func(c *CreateConfig) {
		c.HostConfig = hc
	}
}

func WithCreateConfig(cfg *containerapi.Config) CreateOption {
	return func(c *CreateConfig) {
		c.Config = cfg
	}
}

func WithCreateNetworkConfig(cfg *containerapi.NetworkingConfig) CreateOption {
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
