package container

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cpuguy83/go-docker/container/streamutil"
	"github.com/cpuguy83/go-docker/errdefs"
	"github.com/cpuguy83/go-docker/transport"
	"github.com/cpuguy83/go-docker/version"
)

// AttachOption is used as functional arguments to container attach
type AttachOption func(*AttachConfig)

// AttachConfig holds the options for attaching to a container
type AttachConfig struct {
	Stream     bool
	Stdin      bool
	Stdout     bool
	Stderr     bool
	DetachKeys string
	Logs       bool
}

// AttachIO is used to for providing access to stdio streams of a container
type AttachIO interface {
	Stdin() io.WriteCloser
	Stdout() io.ReadCloser
	Stderr() io.ReadCloser
	Close() error
}

// WithAttachStdin enables stdin on an attach request
func WithAttachStdin(o *AttachConfig) {
	o.Stdin = true
}

// WithAttachStdOut enables stdout on an attach request
func WithAttachStdout(o *AttachConfig) {
	o.Stdout = true
}

// WithAttachStdErr enables stderr on an attach request
func WithAttachStderr(o *AttachConfig) {
	o.Stderr = true
}

// WithAttachStream sets the stream option on an attach request
// When attaching, unless you only want historical data (e.g. setting Logs=true), you probably want this.
func WithAttachStream(o *AttachConfig) {
	o.Stream = true
}

// WithAttachDetachKeys sets the key sequence for detaching from an attach request
func WithAttachDetachKeys(keys string) func(*AttachConfig) {
	return func(o *AttachConfig) {
		o.DetachKeys = keys
	}
}

// Attach attaches to a container's stdio streams.
// You must specify which streams you want to attach to.
// Depending on the container config the streams may not be available for attach.
//
// It is recommend to call `Attach` separately for each stdio stream. This function does support attaching to any/all streams
// in a single request, however the semantics of consuming/blocking the streams is quite a bit more complicated since all i/o
// is multiplexed on a single HTTP stream which can cause one stream to block another if it is not consumed.
//
// Note that unconsumed attach streams can block the stdio of the container process.
//
// It is recommend to instantiate a container object and use the Stdio pipe functions instead of using this.
// This is for advanced use cases only just to expose all the functionality that the Docker API does.
func (s *Service) Attach(ctx context.Context, name string, opts ...AttachOption) (AttachIO, error) {
	var cfg AttachConfig
	cfg.Stream = true
	for _, o := range opts {
		o(&cfg)
	}

	return handleAttach(ctx, s.tr, name, cfg)
}

// TODO: this needs more tests to handle errors cases
func handleAttach(ctx context.Context, tr transport.Doer, name string, cfg AttachConfig) (retAttach *attachIO, retErr error) {
	defer func() {
		if retErr != nil {
			if retAttach != nil {
				retAttach.Close()
			}
		}
	}()

	withAttachRequest := func(req *http.Request) error {
		q := req.URL.Query()
		q.Add("stdin", strconv.FormatBool(cfg.Stdin))
		q.Add("stdout", strconv.FormatBool(cfg.Stdout))
		q.Add("stderr", strconv.FormatBool(cfg.Stderr))
		q.Add("logs", strconv.FormatBool(cfg.Logs))
		q.Add("stream", strconv.FormatBool(cfg.Stream))
		req.URL.RawQuery = q.Encode()
		return nil
	}

	rwc, err := tr.DoRaw(ctx, http.MethodPost, version.Join(ctx, "/containers/"+name+"/attach"), withAttachRequest, transport.WithUpgrade("tcp"))
	if err != nil {
		return nil, err
	}

	var (
		stdin          io.WriteCloser
		stdout, stderr io.ReadCloser
	)

	var isTTY bool
	if cfg.Stdout {
		info, err := handleInspect(ctx, tr, name)
		if err != nil {
			return nil, errdefs.Wrap(err, "error getting container details")
		}
		isTTY = info.Config.Tty
	}

	if isTTY {
		if cfg.Stdout {
			var stdoutW *io.PipeWriter
			stdout, stdoutW = io.Pipe()
			go func() {
				_, err := io.Copy(stdoutW, rwc)
				stdoutW.CloseWithError(err)
				rwc.Close()
			}()
		}
	} else {
		// TODO: This implementation can be a little funky because stderr and stodout
		// are multiplexed on one stream. If the caller doesn't read a full entry
		// from one stream it can block the other stream.
		//
		// TODO: consider using websocket? I'm not sure this actually works correctly
		// in docker.
		var stdoutW, stderrW io.Writer
		closeStdout := func(error) {}
		closeStderr := func(error) {}
		if cfg.Stdout {
			r, w := io.Pipe()
			stdout = r
			stdoutW = w
			closeStdout = func(err error) {
				w.CloseWithError(err)
				closeStderr(err)
			}
		}
		if cfg.Stderr {
			r, w := io.Pipe()
			stderr = r
			stderrW = w
			closeStderr = func(err error) {
				w.CloseWithError(err)
				closeStdout(err)
			}
		}
		if stdoutW != nil && stderrW == nil {
			stderrW = ioutil.Discard
		}
		if stderrW != nil && stdoutW == nil {
			stdoutW = ioutil.Discard
		}
		go func() {
			_, err := streamutil.StdCopy(stdoutW, stderrW, rwc)
			closeStdout(err)
			closeStderr(err)
			rwc.Close()
		}()
	}
	if cfg.Stdin {
		stdin = rwc
	}

	return &attachIO{stdin: stdin, stdout: stdout, stderr: stderr}, nil
}

type attachIO struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func (a *attachIO) Stdin() io.WriteCloser {
	return a.stdin
}

func (a *attachIO) Stdout() io.ReadCloser {
	return a.stdout
}

func (a *attachIO) Stderr() io.ReadCloser {
	return a.stderr
}

func (a *attachIO) Close() error {
	if a.stdin != nil {
		a.stdin.Close()
	}
	if a.stdout != nil {
		a.stdout.Close()
	}
	if a.stderr != nil {
		a.stderr.Close()
	}
	return nil
}

// TODO: Do these pipe calls need to be able to set options like DetachKeys?

// StdinPipe opens a pipe to the container's stdin stream.
// If the container is not configured with `OpenStdin`, this will not work.
func (c *Container) StdinPipe(ctx context.Context) (io.WriteCloser, error) {
	attach, err := handleAttach(ctx, c.tr, c.id, AttachConfig{Stdin: true, Stream: true})
	if err != nil {
		return nil, err
	}
	return attach.Stdin(), nil
}

// StdoutPipe opens a pipe to the container's stdout stream.
func (c *Container) StdoutPipe(ctx context.Context) (io.ReadCloser, error) {
	attach, err := handleAttach(ctx, c.tr, c.id, AttachConfig{Stdout: true, Stream: true})
	if err != nil {
		return nil, err
	}
	return attach.Stdout(), nil
}

// StderrPipe opens a pipe to the container's stderr stream.
func (c *Container) StderrPipe(ctx context.Context) (io.ReadCloser, error) {
	attach, err := handleAttach(ctx, c.tr, c.id, AttachConfig{Stderr: true, Stream: true})
	if err != nil {
		return nil, err
	}
	return attach.Stderr(), nil
}
