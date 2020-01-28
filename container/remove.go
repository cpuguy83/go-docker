package container

import (
	"context"
	"net/http"
	"strconv"

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
		q := req.URL.Query()
		q.Add("force", strconv.FormatBool(cfg.Force))
		q.Add("link", strconv.FormatBool(cfg.RemoveLinks))
		q.Add("v", strconv.FormatBool(cfg.RemoveVolumes))
		req.URL.RawQuery = q.Encode()
		return nil
	}

	resp, err := docker.G(ctx).Do(ctx, http.MethodDelete, "/containers/"+name, withRemoveConfig)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
