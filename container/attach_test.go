package container

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

func TestAttachTTY(t *testing.T) {
	ctx := context.Background()

	c, err := Run(ctx,
		WithRunCreateOption(WithCreateImage("busybox:latest")),
		WithRunCreateOption(WithCreateTTY),
		WithRunCreateOption(WithCreateAttachStdin),
		WithRunCreateOption(WithCreateAttachStdout),
	)
	assert.NilError(t, err)
	defer Remove(ctx, c.ID(), WithRemoveForce)

	attach, err := c.Attach(ctx, WithAttachStdin, WithAttachStdout)
	assert.NilError(t, err)
	defer attach.Close()

	assert.Assert(t, attach.Stdout() != nil)
	assert.Assert(t, attach.Stderr() == nil)
	assert.Equal(t, attach.Stderr(), nil)

	buf := bytes.NewBuffer(nil)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		io.CopyN(buf, attach.Stdout(), 17)
		cancel()
	}()

	n, err := attach.Stdin().Write([]byte("echo hello\n"))
	assert.NilError(t, err)
	assert.Equal(t, n, 11)

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	<-ctx.Done()
	assert.Equal(t, ctx.Err(), context.Canceled)
	assert.Equal(t, buf.Len(), 17)
	assert.Equal(t, buf.String(), "echo hello\r\nhello")
}

func TestAttachNoTTY(t *testing.T) {
	ctx := context.Background()
	c, err := Create(ctx,
		WithCreateImage("busybox:latest"),
		WithCreateAttachStdout,
		WithCreateAttachStderr,
		WithCreateCmd("/bin/sh", "-c", "echo hello; echo world 1>&2"),
	)
	assert.NilError(t, err)
	defer Remove(ctx, c.ID(), WithRemoveForce)

	attach, err := c.Attach(ctx, WithAttachStdout, WithAttachStderr)
	assert.NilError(t, err)
	defer attach.Close()

	assert.Assert(t, attach.Stdout() != nil)
	assert.Assert(t, attach.Stderr() != nil)

	stdoutBuf := bytes.NewBuffer(nil)
	stderrBuf := bytes.NewBuffer(nil)

	chStdout := make(chan struct{})
	go func() {
		io.CopyN(stdoutBuf, attach.Stdout(), 6)
		close(chStdout)
	}()

	chStderr := make(chan struct{})
	go func() {
		io.CopyN(stderrBuf, attach.Stderr(), 6)
		close(chStderr)
	}()

	assert.NilError(t, c.Start(ctx))

	waitFor(t, chStdout, 30*time.Second, "stdout stream")
	assert.Check(t, cmp.Equal(stdoutBuf.String(), "hello\n"))

	waitFor(t, chStderr, 30*time.Second, "stderr stream")
	assert.Check(t, cmp.Equal(stderrBuf.String(), "world\n"))
}

func waitFor(t *testing.T, ch <-chan struct{}, d time.Duration, desc string) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		t.Errorf("timeout waiting for %s", desc)
	case <-ch:
	}
}
