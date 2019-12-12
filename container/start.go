package container

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/docker/docker/errdefs"

	"github.com/pkg/errors"

	"github.com/cpuguy83/go-docker"
)

type StartOption func(*StartConfig)

type StartConfig struct {
	CheckpointID  string
	CheckpointDir string
}

func (c *container) Start(ctx context.Context, opts ...StartOption) error {
	return StartWithClient(ctx, c.client, c.id, opts...)
}

func Start(ctx context.Context, name string, opts ...StartOption) error {
	return StartWithClient(ctx, docker.G(ctx), name, opts...)
}

func StartWithClient(ctx context.Context, client *docker.Client, name string, opts ...StartOption) error {
	if name == "" {
		return errdefs.InvalidParameter(errors.New("must set name value"))
	}

	var cfg StartConfig
	for _, o := range opts {
		o(&cfg)
	}

	withStartConfig := func(req *http.Request) error {
		data, err := json.Marshal(cfg)
		if err != nil {
			return errors.Wrap(err, "error marshaling start config")
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(data))
		return nil
	}

	resp, err := client.Do(ctx, http.MethodPost, "/containers/"+name+"/start", withStartConfig)
	if err != nil {
		return err
	}

	resp.Body.Close()
	return nil
}
