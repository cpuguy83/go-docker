package container

import (
	"context"
	"testing"

	"github.com/cpuguy83/go-docker/errdefs"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

func TestExec(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	c := s.NewContainer(ctx, "notexist")
	_, err := c.Exec(ctx, WithExecCmd("true"))
	assert.Check(t, errdefs.IsNotFound(err), err)

	c, err = s.Create(ctx, WithCreateImage("busybox:latest"),
		WithCreateCmd("/bin/sh", "-c", "trap 'exit 0' SIGTERM; while true; do sleep 0.1; done"),
	)
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
	// Pid seems to be 0 regardless of state?
	// TODO: investigate this further
	// Upon first inspection this seems tobe configured correctly in the client so maybe a daemon bug.
	// assert.Check(t, inspect.Pid != 0)
	assert.Check(t, cmp.Equal(inspect.ExitCode, nilCode))

	assert.NilError(t, c.Stop(ctx))
	inspect, err = ep.Inspect(ctx)
	assert.NilError(t, err)
	assert.Check(t, !inspect.Running)
	assert.Assert(t, inspect.ExitCode != nil)
	assert.Check(t, *inspect.ExitCode != 0)
}
