package container

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
	"github.com/pkg/errors"
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

// TODO: wrap the returned reader in a struct?
// TODO: Provide helper for consuming logs, maybe like daemon/logs does with a channel of discrete log messages?
func (c *Container) Logs(ctx context.Context, opts ...LogsReadOption) (io.ReadCloser, error) {
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

	// Here we do not want to limit the response size since we are returning a log stream, so we perform this manually
	//  instead of with httputil.DoRequest
	resp, err := c.tr.Do(ctx, http.MethodGet, version.Join(ctx, "/container/"+c.id+"/logs"), withLogConfig)
	if err != nil {
		return nil, err
	}

	body := resp.Body
	httputil.LimitResponse(ctx, resp)
	return body, httputil.CheckResponseError(resp)
}
