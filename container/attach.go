package container

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/cpuguy83/go-docker"
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

func (c *container) Attach(ctx context.Context, opts ...AttachOption) (AttachIO, error) {
	attach, err := Attach(ctx, c.id, opts...)
	if err != nil {
		return nil, err
	}
	return attach, nil
}

func Attach(ctx context.Context, name string, opts ...AttachOption) (AttachIO, error) {
	return AttachWithClient(ctx, docker.G(ctx), name, opts...)
}

func AttachWithClient(ctx context.Context, client *docker.Client, name string, opts ...AttachOption) (AttachIO, error) {
	var cfg AttachConfig
	cfg.Stream = true
	for _, o := range opts {
		o(&cfg)
	}

	return handleAttach(ctx, client, name, cfg)
}

func uri(format string, values ...interface{}) string {
	return fmt.Sprintf(format, values...)
}

func handleAttach(ctx context.Context, client *docker.Client, name string, cfg AttachConfig) (retAttach *attachIO, retErr error) {
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

	rwc, err := client.DoRaw(ctx, http.MethodPost, "/containers/"+name+"/attach", withAttachRequest)
	if err != nil {
		return nil, err
	}

	var (
		stdin          io.WriteCloser
		stdout, stderr io.ReadCloser
	)

	var isTTY bool
	if cfg.Stdout {
		info, err := Inspect(ctx, name)
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
