package container

import (
	"context"
	"net/http"

	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/transport"
	"github.com/cpuguy83/go-docker/version"
	"github.com/pkg/errors"
)

// KillOption is a functional argument passed to `Kill`, it is used to configure a KillConfig
type KillOption func(*KillConfig)

// KillConfig holds options available for the kill API
type KillConfig struct {
	Signal string
}

// WithKillSignal returns a KillOption that sets the signal to send to the container
func WithKillSignal(signal string) KillOption {
	return func(cfg *KillConfig) {
		cfg.Signal = signal
	}
}

// Kill sends a signal to the container.
// If no signal is provided docker will send the default signal (SIGKILL on Linux) to the container.
func (s *Service) Kill(ctx context.Context, name string, opts ...KillOption) error {
	return handleKill(ctx, s.tr, name, opts...)
}

func handleKill(ctx context.Context, tr transport.Doer, name string, opts ...KillOption) error {
	var cfg KillConfig
	for _, o := range opts {
		o(&cfg)
	}
	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return tr.Do(ctx, http.MethodPost, version.Join(ctx, "/containers/"+name+"/kill"), func(req *http.Request) error {
			q := req.URL.Query()
			q.Add("signal", cfg.Signal)
			req.URL.RawQuery = q.Encode()
			return nil
		})
	})
	if err != nil {
		return errors.Wrap(err, "error sending signal")
	}
	resp.Body.Close()
	return nil
}

// Kill sends a signal to the container
// If no signal is provided docker will send the default signal (SIGKILL on Linux).
func (c *Container) Kill(ctx context.Context, opts ...KillOption) error {
	return handleKill(ctx, c.tr, c.id, opts...)
}
