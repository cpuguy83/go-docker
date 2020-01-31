package container

import (
	"context"
	"net/http"
	"strconv"
)

// RemoveOption is used as functional arguments for container remove
// RemoveOptioni configure a RemoveConfig
type RemoveOption func(*RemoveConfig)

// RemoveConfig holds options for container remove.
type RemoveConfig struct {
	RemoveVolumes bool
	RemoveLinks   bool
	Force         bool
}

// WithRemoveForce is a RemoveOption that enables the force remove option.
// This enables a container to be removed even if it is running.
func WithRemoveForce(o *RemoveConfig) {
	o.Force = true
}

// Remove removes a container.
func (s *Service) Remove(ctx context.Context, name string, opts ...RemoveOption) error {
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

	resp, err := s.tr.Do(ctx, http.MethodDelete, "/containers/"+name, withRemoveConfig)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
