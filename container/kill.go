package container

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/cpuguy83/go-docker/transport"
)

type KillOption func(*KillConfig)

type KillConfig struct {
	Signal string
}

func (s *Service) Kill(ctx context.Context, name string, opts ...KillOption) error {
	return handleKill(ctx, s.tr, name, opts...)
}

func handleKill(ctx context.Context, tr transport.Doer, name string, opts ...KillOption) error {
	var cfg KillConfig
	for _, o := range opts {
		o(&cfg)
	}
	resp, err := tr.Do(ctx, http.MethodPost, "/containers/"+name+"/kill")
	if err != nil {
		return errors.Wrap(err, "error sending kill signal")
	}
	resp.Body.Close()
	return nil
}

func (c *Container) Kill(ctx context.Context, opts ...KillOption) error {
	return handleKill(ctx, c.tr, c.id, opts...)
}
