package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/cpuguy83/go-docker/container/containerapi"
	"github.com/cpuguy83/go-docker/errdefs"
	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
)

// CreateConfig holds the options for creating a container
type CreateConfig struct {
	Spec     Spec
	Name     string
	Platform string
}

// Spec holds all the configuration for the container create API request
type Spec struct {
	containerapi.Config
	HostConfig    containerapi.HostConfig
	NetworkConfig containerapi.NetworkingConfig
}

func WithCreatePlatform(platform string) CreateOption {
	return func(cfg *CreateConfig) {
		cfg.Platform = platform
	}
}

// WithCreateHostConfigOpt allows you to set a function that modifies the
// HostConfig of the container being created.
func WithCreateHostConfigOpt(f func(*containerapi.HostConfig)) CreateOption {
	return func(cfg *CreateConfig) {
		f(&cfg.Spec.HostConfig)
	}
}

// WithCreateNetworkConfigOpt allows you to set a function that modifies the
// NetworkingConfig of the container being created.
func WithCreateNetworkConfigOpt(f func(*containerapi.NetworkingConfig)) CreateOption {
	return func(cfg *CreateConfig) {
		f(&cfg.Spec.NetworkConfig)
	}
}

// WithCreateConfigOpt allows you to set a function that modifies the
// Config of the container being created.
func WithCreateConfigOpt(f func(*containerapi.Config)) CreateOption {
	return func(cfg *CreateConfig) {
		f(&cfg.Spec.Config)
	}
}

// WithCreatePortForwarding adds the specified port forward to the container's
// configuration.
// In this case the 'port' is the container port to forward to the host.
func WithCreatePortForwarding(proto string, port int, hostBindings ...containerapi.PortBinding) CreateOption {
	return func(cfg *CreateConfig) {
		portSpec := fmt.Sprintf("%d/%s", port, proto)

		WithCreateConfigOpt(func(c *containerapi.Config) {
			if c.ExposedPorts == nil {
				c.ExposedPorts = map[string]struct{}{}
			}
			c.ExposedPorts[portSpec] = struct{}{}
		})(cfg)

		WithCreateHostConfigOpt(func(hc *containerapi.HostConfig) {
			bindings := hc.PortBindings
			if bindings == nil {
				bindings = containerapi.PortMap{}
			}

			bindings[fmt.Sprintf("%d/%s", port, proto)] = hostBindings
			hc.PortBindings = bindings
		})(cfg)
	}
}

// Create creates a container using the provided image.
func (s *Service) Create(ctx context.Context, img string, opts ...CreateOption) (*Container, error) {
	c := CreateConfig{
		Spec: Spec{
			Config: containerapi.Config{Image: img},
		},
	}
	for _, o := range opts {
		o(&c)
	}

	if c.Spec.Config.Image == "" {
		c.Spec.Config.Image = img
	}

	withName := func(req *http.Request) error { return nil }
	if c.Name != "" {
		withName = func(req *http.Request) error {
			q := req.URL.Query()
			if q == nil {
				q = url.Values{}
			}

			q.Set("name", c.Name)
			req.URL.RawQuery = q.Encode()
			return nil
		}
	}

	withPlatform := func(req *http.Request) error {
		q := req.URL.Query()
		q.Set("platform", c.Platform)
		req.URL.RawQuery = q.Encode()
		return nil
	}

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, http.MethodPost, version.Join(ctx, "/containers/create"), httputil.WithJSONBody(c.Spec), withName, withPlatform)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errdefs.Wrap(err, "error reading response body")
	}

	var cc containerCreateResponse
	if err := json.Unmarshal(data, &cc); err != nil {
		return nil, errdefs.Wrap(err, "error decoding container create response")
	}

	if cc.ID == "" {
		return nil, fmt.Errorf("empty ID in response: %v", string(data))
	}
	return &Container{id: cc.ID, tr: s.tr}, nil
}

type containerCreateResponse struct {
	ID string `json:"Id"`
}
