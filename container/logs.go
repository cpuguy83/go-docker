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
	Details    bool `json:"details,omitempty"`
}

// TODO: wrap the returned reader in a struct?
// TODO: Provide helper for consuming logs, maybe like daemon/logs does with a channel of discrete log messages?
func (c *Container) Logs(ctx context.Context, opts ...LogsReadOption) (io.ReadCloser, error) {
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
		return nil, err
	}

	body := resp.Body
	httputil.LimitResponse(ctx, resp)
	return body, httputil.CheckResponseError(resp)
}
