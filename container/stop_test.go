package container

import (
	"context"
	"testing"

	"github.com/cpuguy83/go-docker/errdefs"
	"gotest.tools/v3/assert"
)

func TestStop(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	c := s.NewContainer(ctx, "notexist")
	err := c.Stop(ctx)
	assert.Assert(t, errdefs.IsNotFound(err), err)

	c, err = s.Create(ctx, "busybox:latest",
		WithCreateCmd("/bin/sh", "-c", "trap 'exit 1' EXIT; while true; do sleep 0.1; done"),
	)
	defer func() {
		if c != nil {
			assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
		}
	}()
	assert.NilError(t, err)
	assert.NilError(t, c.Start(ctx))
	assert.NilError(t, c.Stop(ctx))

	inspect, err := c.Inspect(ctx)
	assert.NilError(t, err)
	assert.Assert(t, !inspect.State.Running)
}
