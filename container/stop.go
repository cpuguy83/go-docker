package container

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

type StopOption func(*StopConfig)

type StopConfig struct {
	Timeout *time.Duration
}

func (c *Container) Stop(ctx context.Context, opts ...StopOption) error {
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

	resp, err := c.tr.Do(ctx, http.MethodPost, "/containers/"+c.id+"/stop", withQuery)
	if err != nil {
		return err
	}

	resp.Body.Close()
	return nil
}
