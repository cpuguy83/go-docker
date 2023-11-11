package container

import (
	"context"
	"strings"
	"testing"

	"github.com/cpuguy83/go-docker/errdefs"
	"github.com/cpuguy83/go-docker/testutils"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestKill(t *testing.T) {
	t.Parallel()

	s, ctx := newTestService(t, context.Background())

	err := s.Kill(ctx, "notexist"+testutils.GenerateRandomString())
	assert.Check(t, errdefs.IsNotFound(err), err)

	c, err := s.Create(ctx, "busybox:latest", WithCreateName(strings.ToLower(t.Name())), WithCreateTTY, WithCreateCmd("/bin/sh", "-c", "trap 'exit 0' SIGTERM; while true; do usleep 100000; done"))
	assert.NilError(t, err)
	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()
	assert.NilError(t, c.Start(ctx))

	err = c.Kill(ctx, WithKillSignal("FAKESIG"))
	assert.Check(t, errdefs.IsInvalid(err), err)

	err = c.Kill(ctx, WithKillSignal("SIGUSR1"))
	assert.NilError(t, err)

	inspect, err := c.Inspect(ctx)
	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(inspect.State.Running, true))

	err = c.Kill(ctx, WithKillSignal("SIGKILL"))
	assert.NilError(t, err)
	inspect, err = c.Inspect(ctx)
	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(inspect.State.Running, false))

	assert.NilError(t, c.Start(ctx))
	inspect, err = c.Inspect(ctx)
	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(inspect.State.Running, true))

	err = c.Kill(ctx)
	assert.NilError(t, err)
	inspect, err = c.Inspect(ctx)
	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(inspect.State.Running, false))
}
