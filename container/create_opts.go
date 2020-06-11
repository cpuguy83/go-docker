package container

import (
	"github.com/cpuguy83/go-docker/container/containerapi"
)

// CreateOption is used as functional arguments for creating a container
// CreateOptions configure a CreateConfig
type CreateOption func(*CreateConfig)

// WithCreateHostConfig is a CreateOption which sets the HostConfig for the container create spec
func WithCreateHostConfig(hc containerapi.HostConfig) CreateOption {
	return func(c *CreateConfig) {
		c.Spec.HostConfig = hc
	}
}

// WithCreateConfig is a CreateOption which sets the Config for the container create spec
func WithCreateConfig(cfg containerapi.Config) CreateOption {
	return func(c *CreateConfig) {
		c.Spec.Config = cfg
	}
}

// WithCreateNetworkingConfig is a CreateOption which sets the NetworkConfig for the container create spec
func WithCreateNetworkConfig(cfg containerapi.NetworkingConfig) CreateOption {
	return func(c *CreateConfig) {
		c.Spec.NetworkConfig = cfg
	}
}

// WithCreateName is a CreateOption which sets the container's name
func WithCreateName(name string) CreateOption {
	return func(c *CreateConfig) {
		c.Name = name
	}
}

// WithCreateImage is a CreateOption which sets the container image
func WithCreateImage(image string) CreateOption {
	return func(c *CreateConfig) {
		c.Spec.Image = image
	}
}

// WithCreateCmd is a CreateOption which sets the command to run in the container
func WithCreateCmd(cmd ...string) CreateOption {
	return func(c *CreateConfig) {
		c.Spec.Config.Cmd = cmd
	}
}

// WithCreateTTY is a CreateOption which configures the container with a TTY
func WithCreateTTY(cfg *CreateConfig) {
	cfg.Spec.Config.Tty = true
}

// WithCreateAttachStdin is a CreateOption which enables attaching to the container's stdin
func WithCreateAttachStdin(cfg *CreateConfig) {
	cfg.Spec.AttachStdin = true
	cfg.Spec.OpenStdin = true
}

// WithCreateAttachStdinOnce is a CreateOption which enables attaching to the container's one time
func WithCreateStdinOnce(cfg *CreateConfig) {
	cfg.Spec.StdinOnce = true
}

// WithCreateAttachStdout is a CreateOption which enables attaching to the container's stdout
func WithCreateAttachStdout(cfg *CreateConfig) {
	cfg.Spec.AttachStdout = true
}

// WithCreateAttachStderr is a CreateOption which enables attaching to the container's stderr
func WithCreateAttachStderr(cfg *CreateConfig) {
	cfg.Spec.AttachStderr = true
}

// WithCreateLabels is a CreateOption which sets the labels to attach to the container
func WithCreateLabels(labels map[string]string) CreateOption {
	return func(c *CreateConfig) {
		c.Spec.Config.Labels = labels
	}
}
