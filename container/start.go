package container

import (
	"context"
	"net/http"

	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
)

// StartOption is used as functional arguments for container Start
// A StartOption configures a StartConfig
type StartOption func(*StartConfig)

// StartConfig holds configuration options for container start
type StartConfig struct {
	CheckpointID  string
	CheckpointDir string
}

// Start starts a container
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

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.tr.Do(ctx, http.MethodPost, version.Join(ctx, "/containers/"+c.id+"/start"), withStartConfig)
	})
	if err != nil {
		return err
	}

	resp.Body.Close()
	return nil
}
