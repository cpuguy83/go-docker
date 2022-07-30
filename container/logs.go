package container

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
)

type LogsReadOption func(*LogReadConfig)

type LogReadConfig struct {
	ShowStdout bool   `json:"stdout"`
	ShowStderr bool   `json:"stderr"`
	Since      string `json:"since"`
	Until      string `json:"until"`
	Timestamps bool   `json:"timestamps"`
	Follow     bool   `json:"follow"`
	Tail       string `json:"tail"`
	Details    bool   `json:"details,omitempty"`
}

const (
	mediaTypeMultiplexed = "application/vnd.docker.multiplexed-stream"
)

// Logs returns the logs for a container.
// The logs may be a multiplexed stream with both stdout and stderr, in which case you'll need to split the stream using github.com/cpuguy83/go-docker/container/streamutil.StdCopy
// The bool value returned indicates whether the logs are multiplexed or not.
func (c *Container) Logs(ctx context.Context, opts ...LogsReadOption) (io.ReadCloser, bool, error) {
	var cfg LogReadConfig
	for _, o := range opts {
		o(&cfg)
	}

	withLogConfig := func(req *http.Request) error {
		q := req.URL.Query()
		q.Add("follow", strconv.FormatBool(cfg.Follow))
		q.Add("stdout", strconv.FormatBool(cfg.ShowStdout))
		q.Add("stderr", strconv.FormatBool(cfg.ShowStderr))
		q.Add("since", cfg.Since)
		q.Add("until", cfg.Until)
		q.Add("timestamps", strconv.FormatBool(cfg.Timestamps))
		q.Add("tail", cfg.Tail)

		req.URL.RawQuery = q.Encode()
		return nil
	}

	// Here we do not want to limit the response size since we are returning a log stream, so we perform this manually
	//  instead of with httputil.DoRequest
	resp, err := c.tr.Do(ctx, http.MethodGet, version.Join(ctx, "/containers/"+c.id+"/logs"), withLogConfig)
	if err != nil {
		return nil, false, err
	}

	// Starting with api version 1.42, docker should returnn a header with the content-type indicating if the stream is multiplexed.
	// If the api version is lower then we'll need to inspect the container to determine if the stream is multiplexed.
	mux := resp.Header.Get("Content-Type") == mediaTypeMultiplexed
	if !mux && version.LessThan(version.APIVersion(ctx), "1.42") {
		inspect, err := c.Inspect(ctx)
		if err == nil {
			mux = !inspect.Config.Tty
		}
	}

	body := resp.Body
	httputil.LimitResponse(ctx, resp)
	return body, mux, httputil.CheckResponseError(resp)
}
