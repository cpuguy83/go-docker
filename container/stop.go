package container

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// StopOption is used as functional arguments to container stop
// StopOptions configure a StopConfig
type StopOption func(*StopConfig)

// StopConfig holds the options for stopping a container
type StopConfig struct {
	Timeout *time.Duration
}

// WithStopTimeout sets the timeout for a stop request.
// Docker waits up to the timeout duration for the container to respond to the stop signal configured on the container.
// Once the timeout is reached and the container still has not stopped, Docker will forcefully terminate the process.
func WithStopTimeout(dur time.Duration) StopOption {
	return func(cfg *StopConfig) {
		cfg.Timeout = &dur
	}
}

// Stop stops a container
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
