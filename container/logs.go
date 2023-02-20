package container

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/cpuguy83/go-docker/container/streamutil"
	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
)

type LogsReadOption func(*LogReadConfig)

type LogReadConfig struct {
	Since      string         `json:"since"`
	Until      string         `json:"until"`
	Timestamps bool           `json:"timestamps"`
	Follow     bool           `json:"follow"`
	Tail       string         `json:"tail"`
	Details    bool           `json:"details,omitempty"`
	Stdout     io.WriteCloser `json:"-"`
	Stderr     io.WriteCloser `json:"-"`
}

type logReadConfigAPI struct {
	ShowStdout bool `json:"stdout"`
	ShowStderr bool `json:"stderr"`
	LogReadConfig
}

const (
	mediaTypeMultiplexed = "application/vnd.docker.multiplexed-stream"
)

// Logs returns the logs for a container.
// The logs may be a multiplexed stream with both stdout and stderr, in which case you'll need to split the stream using github.com/cpuguy83/go-docker/container/streamutil.StdCopy
// The bool value returned indicates whether the logs are multiplexed or not.
func (c *Container) Logs(ctx context.Context, opts ...LogsReadOption) error {
	var cfg LogReadConfig
	for _, o := range opts {
		o(&cfg)
	}

	cfgAPI := logReadConfigAPI{
		ShowStdout:    cfg.Stdout != nil,
		ShowStderr:    cfg.Stderr != nil,
		LogReadConfig: cfg,
	}

	withLogConfig := func(req *http.Request) error {
		q := req.URL.Query()
		q.Add("follow", strconv.FormatBool(cfgAPI.Follow))
		q.Add("stdout", strconv.FormatBool(cfgAPI.ShowStdout))
		q.Add("stderr", strconv.FormatBool(cfgAPI.ShowStderr))
		q.Add("since", cfgAPI.Since)
		q.Add("until", cfgAPI.Until)
		q.Add("timestamps", strconv.FormatBool(cfgAPI.Timestamps))
		q.Add("tail", cfgAPI.Tail)

		req.URL.RawQuery = q.Encode()
		return nil
	}

	// Here we do not want to limit the response size since we are returning a log stream, so we perform this manually
	//  instead of with httputil.DoRequest
	resp, err := c.tr.Do(ctx, http.MethodGet, version.Join(ctx, "/containers/"+c.id+"/logs"), withLogConfig)
	if err != nil {
		return err
	}

	// Starting with api version 1.42, docker should returnn a header with the content-type indicating if the stream is multiplexed.
	// If the api version is lower then we'll need to inspect the container to determine if the stream is multiplexed.
	mux := resp.Header.Get("Content-Type") == mediaTypeMultiplexed
	if !mux {
		if version.APIVersion(ctx) == "" || version.LessThan(version.APIVersion(ctx), "1.42") {
			inspect, err := c.Inspect(ctx)
			if err == nil {
				mux = !inspect.Config.Tty
			}
		}
	}

	body := resp.Body
	httputil.LimitResponse(ctx, resp)
	if err := httputil.CheckResponseError(resp); err != nil {
		return err
	}

	if mux {
		if cfg.Stdout != nil || cfg.Stderr != nil {
			go func() {
				streamutil.StdCopy(cfg.Stdout, cfg.Stderr, body)
				closeWrite(cfg.Stdout)
				closeWrite(cfg.Stderr)
				body.Close()
			}()
		}
		return nil
	}

	if cfg.Stdout != nil {
		go func() {
			io.Copy(cfg.Stdout, body)
			closeWrite(cfg.Stdout)
			body.Close()
		}()
	}

	if cfg.Stderr != nil {
		go func() {
			io.Copy(cfg.Stderr, body)
			closeWrite(cfg.Stderr)
			body.Close()
		}()
	}

	return nil
}
