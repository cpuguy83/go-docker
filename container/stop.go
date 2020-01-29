package container

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/cpuguy83/go-docker"
)

type StopOption func(*StopConfig)

type StopConfig struct {
	Timeout *time.Duration
}

func Stop(ctx context.Context, name string, opts ...StopOption) error {
	var cfg StopConfig
	for _, o := range opts {
		o(&cfg)
	}

	// TODO: Set timeout based on context?

	withQuery := func(req *http.Request) error {
		if cfg.Timeout != nil {
			q := req.URL.Query()
			q.Set("timeout", strconv.FormatFloat(cfg.Timeout.Seconds(), 'f', 0, 64))
			req.URL.RawQuery = q.Encode()
		}
		return nil
	}

	resp, err := docker.G(ctx).Do(ctx, http.MethodPost, "/containers/"+name+"/stop", withQuery)
	if err != nil {
		return err
	}

	resp.Body.Close()
	return nil
}

func (c *container) Stop(ctx context.Context, opts ...StopOption) error {
	if err := Stop(ctx, c.id, opts...); err != nil {
		return err
	}
	return nil
}
