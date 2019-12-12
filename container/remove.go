package container

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	"github.com/cpuguy83/go-docker"
)

type RemoveOption func(*RemoveConfig)

type RemoveConfig struct {
	RemoveVolumes bool
	RemoveLinks   bool
	Force         bool
}

func WithRemoveForce(o *RemoveConfig) {
	o.Force = true
}

func Remove(ctx context.Context, name string, opts ...RemoveOption) error {
	var cfg RemoveConfig
	for _, o := range opts {
		o(&cfg)
	}

	withRemoveConfig := func(req *http.Request) error {
		data, err := json.Marshal(cfg)
		if err != nil {
			return errors.Wrap(err, "error marshaling container remove config")
		}

		req.Body = ioutil.NopCloser(bytes.NewReader(data))
		return nil
	}

	resp, err := docker.G(ctx).Do(ctx, http.MethodDelete, "/containers/"+name, withRemoveConfig)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
