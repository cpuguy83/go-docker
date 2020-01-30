package container

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cpuguy83/go-docker/transport"
	"github.com/pkg/errors"
)

type AttachOption func(*AttachConfig)

type AttachConfig struct {
	Stream     bool
	Stdin      bool
	Stdout     bool
	Stderr     bool
	DetachKeys string
	Logs       bool
}

type AttachIO interface {
	Stdin() io.WriteCloser
	Stdout() io.ReadCloser
	Stderr() io.ReadCloser
	Close() error
}

func WithAttachStdin(o *AttachConfig) {
	o.Stdin = true
}

func WithAttachStdout(o *AttachConfig) {
	o.Stdout = true
}

func WithAttachStderr(o *AttachConfig) {
	o.Stderr = true
}

func WithAttachStream(o *AttachConfig) {
	o.Stream = true
}

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
func (s *Service) Attach(ctx context.Context, name string, opts ...AttachOption) (AttachIO, error) {
	var cfg AttachConfig
	cfg.Stream = true
	for _, o := range opts {
		o(&cfg)
	}

	return handleAttach(ctx, s.tr, name, cfg)
}

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

	rwc, err := tr.DoRaw(ctx, http.MethodPost, "/containers/"+name+"/attach", withAttachRequest)
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
			return nil, errors.Wrap(err, "error getting container details")
		}
		isTTY = info.Config.Tty
	}

	if isTTY {
		if cfg.Stdout {
			var stdoutW io.WriteCloser
			stdout, stdoutW = io.Pipe()
			go io.Copy(stdoutW, rwc)
		}
	} else {
		// TODO: This implementation can be a little funky because stderr and stodout
		// are multiplexed on one stream. If the caller doesn't read a full entry
		// from one stream it can block the other stream.
		//
		// TODO: consider using websocket? I'm not sure this actually works correctly
		// in docker.
		var stdoutW, stderrW io.Writer
		if cfg.Stdout {
			stdout, stdoutW = io.Pipe()
		}
		if cfg.Stderr {
			stderr, stderrW = io.Pipe()
		}
		if stdoutW != nil && stderrW == nil {
			stderrW = ioutil.Discard
		}
		if stderrW != nil && stdoutW == nil {
			stdoutW = ioutil.Discard
		}
		go stdCopy(stdoutW, stderrW, rwc)
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

func (c *Container) StdinPipe(ctx context.Context) (io.WriteCloser, error) {
	attach, err := handleAttach(ctx, c.tr, c.id, AttachConfig{Stdin: true, Stream: true})
	if err != nil {
		return nil, err
	}
	return attach.Stdin(), nil
}

func (c *Container) StdoutPipe(ctx context.Context) (io.ReadCloser, error) {
	attach, err := handleAttach(ctx, c.tr, c.id, AttachConfig{Stdout: true, Stream: true})
	if err != nil {
		return nil, err
	}
	return attach.Stdout(), nil
}

func (c *Container) StderrPipe(ctx context.Context) (io.ReadCloser, error) {
	attach, err := handleAttach(ctx, c.tr, c.id, AttachConfig{Stderr: true, Stream: true})
	if err != nil {
		return nil, err
	}
	return attach.Stderr(), nil
}
