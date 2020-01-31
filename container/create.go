package container

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/cpuguy83/go-docker/container/containerapi"
	"github.com/pkg/errors"
)

const DefaultCreateDecodeLimitBytes = 64 * 1024

type CreateConfig struct {
	Config           *containerapi.Config
	DecodeLimitBytes int64
	HostConfig       *containerapi.HostConfig
	NetworkConfig    *containerapi.NetworkingConfig
	Name             string
}

func (s *Service) Create(ctx context.Context, opts ...CreateOption) (*Container, error) {
	c := CreateConfig{
		Config:           &containerapi.Config{},
		HostConfig:       &containerapi.HostConfig{},
		NetworkConfig:    &containerapi.NetworkingConfig{},
		DecodeLimitBytes: DefaultCreateDecodeLimitBytes,
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
	cw := &containerConfigWrapper{Config: c.Config, HostConfig: c.HostConfig, NetworkingConfig: c.NetworkConfig}

	resp, err := s.tr.Do(ctx, http.MethodPost, "/containers/create", withJSONBody(cw), withName)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body := io.LimitReader(resp.Body, c.DecodeLimitBytes)

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
	return &Container{id: cc.ID, tr: s.tr}, nil
}

type containerConfigWrapper struct {
	*containerapi.Config
	HostConfig       *containerapi.HostConfig
	NetworkingConfig *containerapi.NetworkingConfig
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
