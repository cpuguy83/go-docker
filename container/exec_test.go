package container

import (
	"bufio"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/cpuguy83/go-docker/errdefs"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestExec(t *testing.T) {
	t.Parallel()

	s, ctx := newTestService(t, context.Background())

	c := s.NewContainer(ctx, "notexist")
	_, err := c.Exec(ctx, WithExecCmd("true"))
	assert.Check(t, errdefs.IsNotFound(err), err)

	c, err = s.Create(ctx, "busybox:latest",
		WithCreateCmd("/bin/sh", "-c", "trap 'exit 0' SIGTERM; while true; do sleep 0.1; done"),
	)
	assert.NilError(t, err)
	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	_, err = c.Exec(ctx, WithExecCmd("/bin/echo", "hello"))
	assert.Check(t, errdefs.IsConflict(err))

	assert.NilError(t, c.Start(ctx))

	ep, err := c.Exec(ctx, WithExecCmd("/bin/sh", "-c", "trap 'exit 0' SIGTERM; while true; do sleep 0.1; done"))
	assert.NilError(t, err)

	inspect, err := ep.Inspect(ctx)
	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(inspect.ID, ep.ID()))
	assert.Check(t, cmp.Equal(inspect.ContainerID, c.ID()))
	assert.Check(t, !inspect.Running)
	assert.Check(t, cmp.Equal(inspect.Pid, 0))
	var nilCode *int
	assert.Check(t, cmp.Equal(inspect.ExitCode, nilCode))

	assert.NilError(t, ep.Start(ctx))

	inspect, err = ep.Inspect(ctx)
	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(inspect.ID, ep.ID()))
	assert.Check(t, cmp.Equal(inspect.ContainerID, c.ID()))
	assert.Check(t, inspect.Running)
	assert.Check(t, inspect.Pid != 0)
	assert.Check(t, cmp.Equal(inspect.ExitCode, nilCode))

	assert.NilError(t, c.Stop(ctx))
	inspect, err = ep.Inspect(ctx)
	assert.NilError(t, err)
	assert.Check(t, !inspect.Running)
	assert.Assert(t, inspect.ExitCode != nil)
	assert.Check(t, *inspect.ExitCode != 0)

	assert.NilError(t, c.Start(ctx))

	r, w := io.Pipe()
	defer r.Close()

	ep, err = c.Exec(ctx, WithExecCmd("cat"), func(cfg *ExecConfig) {
		cfg.Stdin = io.NopCloser(strings.NewReader("hello\n"))
		cfg.Stdout = w
		cfg.Stderr = w
	})
	assert.NilError(t, err)

	err = ep.Start(ctx)
	assert.NilError(t, err)

	line, _ := bufio.NewReader(r).ReadString('\n')
	assert.Equal(t, line, "hello\n")
}
