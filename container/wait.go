package container

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// WaitCondition is a type used to specify a container state for which
// to wait.
type WaitCondition string

// Possible WaitCondition Values.
//
// WaitConditionNotRunning (default) is used to wait for any of the non-running
// states: "created", "exited", "dead", "removing", or "removed".
//
// WaitConditionNextExit is used to wait for the next time the state changes
// to a non-running state. If the state is currently "created" or "exited",
// this would cause Wait() to block until either the container runs and exits
// or is removed.
//
// WaitConditionRemoved is used to wait for the container to be removed.
const (
	WaitConditionNotRunning WaitCondition = "not-running"
	WaitConditionNextExit   WaitCondition = "next-exit"
	WaitConditionRemoved    WaitCondition = "removed"

	// DefaultWaitDecodeLimitBytes is the default max size that will be read from a container wait response.
	// This value is used if no value is set on CreateConfig.
	DefaultWaitDecodeLimitBytes = 64 * 1024
)

// WaitConfig holds the options for waiting on a container
type WaitConfig struct {
	Condition        WaitCondition
	DecodeLimitBytes int64
}

// WaitOption is used as functional arguments to container wait
// WaitOptions configure a WaitConfig
type WaitOption func(*WaitConfig)

// ExitStatus is used to report information about a container exit
// It is used by container.Wait.
type ExitStatus interface {
	ExitCode() int
}

type waitStatus struct {
	StatusCode int
	Err        *struct {
		Message string
	} `json:"Error"`
}

func (s *waitStatus) ExitCode() int {
	return s.StatusCode
}

func WithWaitCondition(cond WaitCondition) WaitOption {
	return func(cfg *WaitConfig) {
		cfg.Condition = cond
	}
}

// Wait waits on the container to meet the provided wait condition.
func (c *Container) Wait(ctx context.Context, opts ...WaitOption) (ExitStatus, error) {
	var cfg WaitConfig
	for _, o := range opts {
		o(&cfg)
	}

	if cfg.DecodeLimitBytes == 0 {
		cfg.DecodeLimitBytes = DefaultWaitDecodeLimitBytes
	}

	resp, err := c.tr.Do(ctx, http.MethodPost, "/containers/"+c.id+"/wait", func(req *http.Request) error {
		q := req.URL.Query()
		q.Add("condition", string(cfg.Condition))
		req.URL.RawQuery = q.Encode()
		return nil
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ws waitStatus
	if err := json.NewDecoder(io.LimitReader(resp.Body, cfg.DecodeLimitBytes)).Decode(&ws); err != nil {
		return nil, errors.Wrap(err, "could not decode resp")
	}

	if ws.Err != nil && ws.Err.Message != "" {
		return &ws, errors.New(ws.Err.Message)
	}
	return &ws, nil
}
