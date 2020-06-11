package container

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/cpuguy83/go-docker/httputil"

	"github.com/cpuguy83/go-docker/version"

	"github.com/pkg/errors"

	"github.com/cpuguy83/go-docker/container/containerapi"
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

// Create creates a container.
// You must specify a CreateOption which sets the image to use (e.g. WithCreateImage) otherwise the API will (should)
// return an error.
// All other options are truly optional.
//
// TODO: Should "image" be moved to a dedicated function argument?
func (s *Service) Create(ctx context.Context, opts ...CreateOption) (*Container, error) {
	c := CreateConfig{
		Spec: Spec{},
	}
	for _, o := range opts {
		o(&c)
	}

	withName := func(req *http.Request) error { return nil }
	if c.Name != "" {
		withName = func(req *http.Request) error {
			q := req.URL.Query()
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
		return nil, errors.Wrap(err, "error reading response body")
	}

	var cc containerCreateResponse
	if err := json.Unmarshal(data, &cc); err != nil {
		return nil, errors.Wrap(err, "error decoding container create response")
	}

	if cc.ID == "" {
		return nil, errors.Errorf("empty ID in response: %v", string(data))
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
			return errors.Wrap(err, "error marshaling json body")
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(data))
		if req.Header == nil {
			req.Header = http.Header{}
		}
		req.Header.Set("Content-Type", "application/json")
		return nil
	}
}
