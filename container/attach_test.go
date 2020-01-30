package container

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/cpuguy83/go-docker/testutils"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

func TestContainerAttachTTY(t *testing.T) {
	ctx := context.Background()
	tr := testutils.NewDefaultTestTransport(t)
	s := NewService(tr)

	c, err := s.Create(ctx,
		WithCreateImage("busybox:latest"),
		WithCreateTTY,
		WithCreateAttachStdin,
		WithCreateAttachStdout,
	)
	assert.NilError(t, err)
	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	stdout, err := c.StdoutPipe(ctx)
	assert.NilError(t, err)
	defer stdout.Close()

	assert.Assert(t, c.Start(ctx), "failed to start container")

	expected := "/ # echo hello\r\nhello\r\n/ # "

	stdin, err := c.StdinPipe(ctx)
	assert.NilError(t, err)
	defer stdin.Close()

	chErr := make(chan error, 1)
	go func() {
		_, err := stdin.Write([]byte("echo hello\n"))
		chErr <- err
	}()

	buf := bytes.NewBuffer(nil)
	_, err = io.CopyN(buf, stdout, int64(len(expected)))
	assert.Check(t, <-chErr)
	assert.Check(t, err)
	assert.Check(t, cmp.Equal(buf.String(), expected))
}

func TestContainerAttachNoTTY(t *testing.T) {
	ctx := context.Background()
	tr := testutils.NewDefaultTestTransport(t)
	s := NewService(tr)

	c, err := s.Create(ctx,
		WithCreateImage("busybox:latest"),
		WithCreateAttachStdout,
		WithCreateAttachStderr,
		WithCreateCmd("/bin/sh", "-c", "echo hello; >&2 echo world"),
	)
	assert.NilError(t, err)
	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	stdout, err := c.StdoutPipe(ctx)
	assert.NilError(t, err)
	defer stdout.Close()

	stderr, err := c.StderrPipe(ctx)
	assert.NilError(t, err)
	defer stderr.Close()

	assert.Assert(t, c.Start(ctx))

	outBuff := bytes.NewBuffer(nil)
	chErr := make(chan error, 2)
	go func() {
		_, err := io.CopyN(outBuff, stdout, 6)
		chErr <- err
	}()
	errBuff := bytes.NewBuffer(nil)
	go func() {
		_, err := io.CopyN(errBuff, stderr, 6)
		chErr <- err
	}()

	assert.Check(t, <-chErr)
	assert.Check(t, <-chErr)
	assert.Check(t, cmp.Equal(outBuff.String(), "hello\n"))
	assert.Check(t, cmp.Equal(errBuff.String(), "world\n"))
}
