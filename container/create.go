package container

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/cpuguy83/go-docker/container/containerapi"
	"github.com/cpuguy83/go-docker/errdefs"
	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
)

// CreateConfig holds the options for creating a container
type CreateConfig struct {
	Spec Spec
	Name string
}

// Spec holds all the configuration for the container create API request
type Spec struct {
	containerapi.Config
	HostConfig    containerapi.HostConfig
	NetworkConfig containerapi.NetworkingConfig
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

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, http.MethodPost, version.Join(ctx, "/containers/create"), withJSONBody(c.Spec), withName)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
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

func withJSONBody(v interface{}) func(req *http.Request) error {
	return func(req *http.Request) error {
		data, err := json.Marshal(v)
		if err != nil {
			return errdefs.Wrap(err, "error marshaling json body")
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(data))
		if req.Header == nil {
			req.Header = http.Header{}
		}
		req.Header.Set("Content-Type", "application/json")
		return nil
	}
}
