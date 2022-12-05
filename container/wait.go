package container

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/cpuguy83/go-docker/errdefs"
	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
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
	Condition WaitCondition
}

// WaitOption is used as functional arguments to container wait
// WaitOptions configure a WaitConfig
type WaitOption func(*WaitConfig)

// ExitStatus is used to report information about a container exit
// It is used by container.Wait.
type ExitStatus interface {
	ExitCode() (int, error)
}

type waitStatus struct {
	mu         sync.Mutex
	ready      bool
	StatusCode int
	err        error
	Err        *struct {
		Message string
	} `json:"Error"`
}

func (s *waitStatus) ExitCode() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.StatusCode, s.err
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

	if version.LessThan(version.APIVersion(ctx), "1.30") {
		// Before 1.30:
		//   - wait condition is not supported
		//   - The API blocks until wait is completed
		//
		// On 2nd point above, this would require running the request in a goroutine.
		// Not difficult but for now just return an error.
		return nil, errdefs.NotImplemented("container wait requires API version 1.30 or higher")
	}
	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.tr.Do(ctx, http.MethodPost, version.Join(ctx, "/containers/"+c.id+"/wait"), func(req *http.Request) error {
			q := req.URL.Query()
			q.Add("condition", string(cfg.Condition))
			req.URL.RawQuery = q.Encode()
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	ws := &waitStatus{}
	ws.mu.Lock()

	go func() {
		defer ws.mu.Unlock()
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&ws); err != nil {
			ws.err = fmt.Errorf("could not decode response: %w", err)
			ws.StatusCode = -1
			return
		}

		if ws.Err != nil && ws.Err.Message != "" {
			ws.err = errors.New(ws.Err.Message)
		}
	}()

	return ws, nil
}
