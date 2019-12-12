package container

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

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
	return CreateWithClient(ctx, docker.G(ctx), opts...)
}

func CreateWithClient(ctx context.Context, client *docker.Client, opts ...CreateOption) (Container, error) {
	c := CreateConfig{
		Config:        &containertypes.Config{},
		HostConfig:    &containertypes.HostConfig{},
		NetworkConfig: &network.NetworkingConfig{},
	}
	for _, o := range opts {
		o(&c)
	}
	withName := func(*http.Request) error { return nil }
	if c.Name != "" {
		docker.WithQueryValue("name", c.Name)
	}

	resp, err := client.Do(ctx, http.MethodPost, "/containers/create", docker.WithJSONBody(c), withName)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body := io.LimitReader(resp.Body, 64*1024)

	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	var cc containerCreateResponse
	if err := json.Unmarshal(data, &cc); err != nil {
		return nil, errors.Wrap(err, "error decoding container create response")
	}

	if cc.ID == "" {
		return nil, errors.Errorf("empty ID in response: %v", string(data))
	}
	return &container{id: cc.ID, client: client}, nil
}

type containerCreateResponse struct {
	ID string `json:"Id"`
}
