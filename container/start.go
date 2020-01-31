package container

import (
	"context"
	"net/http"
)

type StartOption func(*StartConfig)

type StartConfig struct {
	CheckpointID  string
	CheckpointDir string
}

func (c *Container) Start(ctx context.Context, opts ...StartOption) error {
	var cfg StartConfig
	for _, o := range opts {
		o(&cfg)
	}

	withStartConfig := func(req *http.Request) error {
		q := req.URL.Query()
		if cfg.CheckpointID != "" {
			q.Add("checkpoint", cfg.CheckpointID)
		}
		if cfg.CheckpointDir != "" {
			q.Add("checkpoint-dir", cfg.CheckpointDir)
		}
		req.URL.RawQuery = q.Encode()
		return nil
	}

	resp, err := c.tr.Do(ctx, http.MethodPost, "/containers/"+c.id+"/start", withStartConfig)
	if err != nil {
		return err
	}

	resp.Body.Close()
	return nil
}
