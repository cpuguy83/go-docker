package container

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	"github.com/cpuguy83/go-docker"
)

type LogsReadOption func(*LogReadConfig)

type LogReadConfig struct {
	ShowStdout bool
	ShowStderr bool
	Since      string
	Until      string
	Timestamps bool
	Follow     bool
	Tail       string
	Details    bool
}

func (c *container) Logs(ctx context.Context, opts ...LogsReadOption) (io.ReadCloser, error) {
	return LogsWithClient(ctx, c.client, c.id, opts...)
}

func Logs(ctx context.Context, name string, opts ...LogsReadOption) (io.ReadCloser, error) {
	return LogsWithClient(ctx, docker.G(ctx), name, opts...)
}

// TODO: wrap the returned reader in a struct?
// TODO: Provide helper for consuming logs, maybe like daemon/logs does with a channel of discrete log messages?
func LogsWithClient(ctx context.Context, client *docker.Client, name string, opts ...LogsReadOption) (io.ReadCloser, error) {
	var cfg LogReadConfig
	for _, o := range opts {
		o(&cfg)
	}

	withLogConfig := func(req *http.Request) error {
		data, err := json.Marshal(&cfg)
		if err != nil {
			return errors.Wrap(err, "error encoding log read config")
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(data))
		return nil
	}

	resp, err := client.Do(ctx, http.MethodGet, "/container/"+name+"/logs", withLogConfig)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
