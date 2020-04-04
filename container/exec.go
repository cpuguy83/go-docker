package container

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/cpuguy83/go-docker/transport"
)

// DefaultExecDecodeLimitBytes is the default max size that will be read from a container create response.
// This value is used if no value is set on CreateConfig.
const DefaultExecDecodeLimitBytes = 16 * 1024

// ExecProcess represents an "Exec"'d process in a container.
type ExecProcess struct {
	id string
	tr transport.Doer
}

// ExecConfig holds all the options for creating a new process in a container
type ExecConfig struct {
	DeclodeLimitBytes int64    `json:"-"`
	User              string   // User that will run the command
	Privileged        bool     // Is the container in privileged mode
	Tty               bool     // Attach standard streams to a tty.
	AttachStdin       bool     // Attach the standard input, makes possible user interaction
	AttachStderr      bool     // Attach the standard error
	AttachStdout      bool     // Attach the standard output
	Detach            bool     // Execute in detach mode
	DetachKeys        string   // Escape keys for detach
	Env               []string // Environment variables
	WorkingDir        string   // Working directory
	Cmd               []string // Execution commands and args
}

// ExecOption is used as functional arguments to configure an ExecConfig
type ExecOption func(config *ExecConfig)

// WithExecCmd is an ExecOption that sets the command to execute in the container
func WithExecCmd(cmd ...string) ExecOption {
	return func(cfg *ExecConfig) {
		cfg.Cmd = cmd
	}
}

type execCreateResponse struct {
	ID string
}

// Exec creates a new process in the container
// The process is not actually started until `Start` is called
func (c *Container) Exec(ctx context.Context, opts ...ExecOption) (*ExecProcess, error) {
	var cfg ExecConfig
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.DeclodeLimitBytes == 0 {
		cfg.DeclodeLimitBytes = DefaultExecDecodeLimitBytes
	}

	resp, err := c.tr.Do(ctx, http.MethodPost, "/containers/"+c.id+"/exec", withJSONBody(cfg))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		return nil, errors.Errorf("expected response Content-Type=application/json for exec create response, got: %s", ct)
	}

	var id execCreateResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, cfg.DeclodeLimitBytes)).Decode(&id); err != nil {
		return nil, errors.Wrap(err, "error decoding exec create response body")
	}

	return &ExecProcess{id: id.ID, tr: c.tr}, nil
}

// ExecStartOption is used as functional arguments to configure an ExecStartConfig
type ExecStartOption func(config *ExecStartConfig)

// ExecStartConfig holds all the options for starting a new process in a container
type ExecStartConfig struct {
	Detach bool
}

// Start starts the exec process
//
// TODO: The API for exec is kind of weird... start is used both for attach and start. If attach is used it must hijack
// For now I would like to only support start and look at adding an API to the Docker API for a more generic attach.
func (e *ExecProcess) Start(ctx context.Context, opts ...ExecStartOption) error {
	var cfg ExecStartConfig
	// detach otherwise the API will basically be async.
	// For instance, if you call start, it returns successful, then inspect, you can end up in a race where pid can be
	// 0 still.
	cfg.Detach = true
	for _, o := range opts {
		o(&cfg)
	}

	resp, err := e.tr.Do(ctx, http.MethodPost, "/exec/"+e.id+"/start", withJSONBody(cfg))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// ExecInspectConfig holds all the options for inspecting an exec process
type ExecInspectConfig struct {
	DecodeLimitBytes int64
}

// ExecInspectOption is used as functional arguments for configuring an ExecInspectConfig
type ExecInspectOption func(*ExecInspectConfig)

// ExecInspect holds detailed information about an exec'd process.
type ExecInspect struct {
	ID            string
	Running       bool
	ExitCode      *int               `json:",omitempty"`
	ProcessConfig *ExecProcessConfig `json:",omitempty"`
	OpenStdin     bool
	OpenStderr    bool
	OpenStdout    bool
	CanRemove     bool
	ContainerID   string
	DetachKeys    []byte
	Pid           int
}

// ExecProcessConfig holds information about the exec process
// running on the host.
type ExecProcessConfig struct {
	Tty        bool     `json:"tty"`
	Entrypoint string   `json:"entrypoint"`
	Cmd        []string `json:"arguments"`
	Privileged *bool    `json:"privileged,omitempty"`
	User       string   `json:"user,omitempty"`
}

func (e *ExecProcess) ID() string {
	return e.id
}

// Inspect returns detailed information about the exec'd process.
func (e *ExecProcess) Inspect(ctx context.Context, opts ...ExecInspectOption) (ExecInspect, error) {
	var cfg ExecInspectConfig
	for _, o := range opts {
		o(&cfg)
	}

	if cfg.DecodeLimitBytes == 0 {
		cfg.DecodeLimitBytes = DefaultExecDecodeLimitBytes
	}

	var inspect ExecInspect
	resp, err := e.tr.Do(ctx, http.MethodGet, "/exec/"+e.id+"/json")
	if err != nil {
		return inspect, err
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		return inspect, errors.Errorf("expected response Content-Type=application/json for exec create response, got: %s", ct)
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, cfg.DecodeLimitBytes)).Decode(&inspect); err != nil {
		return inspect, errors.Wrap(err, "error decoding exec inspect response")
	}
	return inspect, nil
}

// ExecResizeConfig holds the options for resizing an exec TTY
type ExecResizeConfig struct {
	Width  int
	Height int
}

// Resize resizes the exec processes TTY
func (e *ExecProcess) Resize(ctx context.Context, cfg ExecResizeConfig) error {
	resp, err := e.tr.Do(ctx, http.MethodPost, "/exec/"+e.id+"/resize", func(req *http.Request) error {
		q := req.URL.Query()
		q.Add("w", strconv.Itoa(cfg.Width))
		q.Add("h", strconv.Itoa(cfg.Height))
		req.URL.RawQuery = q.Encode()
		return nil
	})
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
