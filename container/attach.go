package container

import (
	"context"
	"io"

	"github.com/cpuguy83/go-docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pkg/errors"
)

type AttachOption func(*types.ContainerAttachOptions)

type AttachIO interface {
	Stdin() io.WriteCloser
	Stdout() io.ReadCloser
	Stderr() io.ReadCloser
	Close() error
}

func WithAttachStdin(o *types.ContainerAttachOptions) {
	o.Stdin = true
}

func WithAttachStdout(o *types.ContainerAttachOptions) {
	o.Stdout = true
}

func WithAttachStderr(o *types.ContainerAttachOptions) {
	o.Stderr = true
}

func (c *container) Attach(ctx context.Context, opts ...AttachOption) (AttachIO, error) {
	return Attach(ctx, c.id, opts...)
}

func Attach(ctx context.Context, name string, opts ...AttachOption) (AttachIO, error) {
	var cfg types.ContainerAttachOptions
	for _, o := range opts {
		o(&cfg)
	}

	cfg.Stream = true

	hr, err := docker.G(ctx).ContainerAttach(ctx, name, cfg)
	if err != nil {
		return nil, err
	}

	var (
		stdout, stderr io.ReadCloser
		stdin          io.WriteCloser
	)

	info, err := Inspect(ctx, name)
	if err != nil {
		return nil, errors.Wrap(err, "error getting container details")
	}

	if info.Config.Tty {
		if cfg.Stdout {
			var stdoutW io.WriteCloser
			stdout, stdoutW = io.Pipe()
			go io.Copy(stdoutW, hr.Reader)
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
		go stdcopy.StdCopy(stdoutW, stderrW, hr.Reader)
	}
	if cfg.Stdin {
		stdin = hr.Conn
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
