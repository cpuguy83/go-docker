package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/cpuguy83/go-docker/container/streamutil"
	"github.com/cpuguy83/go-docker/errdefs"
	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/transport"
	"github.com/cpuguy83/go-docker/version"
)

// DefaultExecDecodeLimitBytes is the default max size that will be read from a container create response.
// This value is used if no value is set on CreateConfig.
const DefaultExecDecodeLimitBytes = 16 * 1024

// ExecProcess represents an "Exec"'d process in a container.
type ExecProcess struct {
	id string
	tr transport.Doer

	stdin  io.ReadCloser
	stdout io.WriteCloser
	stderr io.WriteCloser
}

// ExecConfig holds all the options for creating a new process in a container
type execConfig struct {
	User         string   // User that will run the command
	Privileged   bool     // Is the container in privileged mode
	Tty          bool     // Attach standard streams to a tty.
	AttachStdin  bool     // Attach the standard input, makes possible user interaction
	AttachStderr bool     // Attach the standard error
	AttachStdout bool     // Attach the standard output
	Detach       bool     // Execute in detach mode
	DetachKeys   string   // Escape keys for detach
	Env          []string // Environment variables
	WorkingDir   string   // Working directory
	Cmd          []string // Execution commands and args
}

type ExecConfig struct {
	Cmd        []string
	User       string
	Privileged bool
	Tty        bool
	Env        []string
	WorkingDir string
	DetachKeys string

	Stdin  io.ReadCloser
	Stdout io.WriteCloser
	Stderr io.WriteCloser
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
//
// Note: the exec API sucks.
// We must know ahead of time that the process will be attached or detached.
// Then it can only be attached when calling the start API.
// So this diverges significantly from the container API.
func (c *Container) Exec(ctx context.Context, opts ...ExecOption) (*ExecProcess, error) {
	var cfg ExecConfig
	for _, o := range opts {
		o(&cfg)
	}

	if len(cfg.Cmd) == 0 {
		return nil, errdefs.Invalid("no command specified")
	}

	cfgApi := execConfig{
		AttachStdin:  cfg.Stdin != nil,
		AttachStdout: cfg.Stdout != nil,
		AttachStderr: cfg.Stderr != nil,
		Cmd:          cfg.Cmd,
		Detach:       cfg.Stdin == nil && cfg.Stdout == nil && cfg.Stderr == nil,
		Env:          cfg.Env,
		Privileged:   cfg.Privileged,
		Tty:          cfg.Tty,
		User:         cfg.User,
		WorkingDir:   cfg.WorkingDir,
	}

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.tr.Do(ctx, http.MethodPost, version.Join(ctx, "/containers/"+c.id+"/exec"), httputil.WithJSONBody(cfgApi))
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		return nil, fmt.Errorf("expected response Content-Type=application/json for exec create response, got: %s", ct)
	}

	var id execCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&id); err != nil {
		return nil, errdefs.Wrap(err, "error decoding exec create response body")
	}

	return &ExecProcess{id: id.ID, tr: c.tr, stdin: cfg.Stdin, stdout: cfg.Stdout, stderr: cfg.Stderr}, nil
}

// ExecStartOption is used as functional arguments to configure an ExecStartConfig
type ExecStartOption func(config *ExecStartConfig)

// ExecStartConfig holds all the options for starting a new process in a container
type ExecStartConfig struct {
}

type apiExecStartConfig struct {
	Detach bool
}

type closeReader interface {
	CloseRead() error
}

type closeWriter interface {
	CloseWrite() error
}

func closeWrite(c io.WriteCloser) error {
	if c == nil {
		return nil
	}
	if cw, ok := c.(closeWriter); ok {
		return cw.CloseWrite()
	}
	return c.Close()
}

func closeRead(c io.ReadCloser) error {
	if c == nil {
		return nil
	}
	if cr, ok := c.(closeReader); ok {
		return cr.CloseRead()
	}
	return c.Close()
}

func (e *ExecProcess) shouldAttach() bool {
	return e.stdin != nil || e.stdout != nil || e.stderr != nil
}

// Start starts the exec process
func (e *ExecProcess) Start(ctx context.Context, opts ...ExecStartOption) error {
	var cfg ExecStartConfig
	for _, o := range opts {
		o(&cfg)
	}

	if !e.shouldAttach() {
		aCfg := apiExecStartConfig{Detach: true}
		resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
			return e.tr.Do(ctx, http.MethodPost, version.Join(ctx, "/exec/"+e.id+"/start"), httputil.WithJSONBody(aCfg))
		})
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	}

	rwc, err := e.tr.DoRaw(ctx, http.MethodPost, version.Join(ctx, "/exec/"+e.id+"/start"), httputil.WithJSONBody(apiExecStartConfig{}), transport.WithUpgrade("tcp"))
	if err != nil {
		return err
	}

	go func() {
		streamutil.StdCopy(e.stdout, e.stderr, rwc)
		closeRead(rwc)
		closeWrite(e.stdout)
		closeWrite(e.stderr)
	}()

	if e.stdin != nil {
		go func() {
			io.Copy(rwc, e.stdin)
			closeWrite(rwc)
			closeRead(e.stdin)
		}()
	}
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
	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return e.tr.Do(ctx, http.MethodGet, version.Join(ctx, "/exec/"+e.id+"/json"))
	})
	if err != nil {
		return inspect, err
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		return inspect, fmt.Errorf("expected response Content-Type=application/json for exec create response, got: %s", ct)
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, cfg.DecodeLimitBytes)).Decode(&inspect); err != nil {
		return inspect, errdefs.Wrap(err, "error decoding exec inspect response")
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
	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return e.tr.Do(ctx, http.MethodPost, version.Join(ctx, "/exec/"+e.id+"/resize"), func(req *http.Request) error {
			q := req.URL.Query()
			q.Add("w", strconv.Itoa(cfg.Width))
			q.Add("h", strconv.Itoa(cfg.Height))
			req.URL.RawQuery = q.Encode()
			return nil
		})
	})
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
