package container

import (
	"context"
	"testing"

	"github.com/cpuguy83/go-docker/errdefs"
	"gotest.tools/v3/assert"
)

func TestStart(t *testing.T) {
	s, ctx := newTestService(t, context.Background())

	c := s.NewContainer(ctx, "notexist")
	err := c.Start(ctx)
	assert.Assert(t, errdefs.IsNotFound(err), err)

	c, err = s.Create(ctx, "busybox:latest", WithCreateCmd("top"))
	defer func() {
		if c != nil {
			assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
		}
	}()
	assert.NilError(t, err)
	assert.NilError(t, c.Start(ctx))
	assert.NilError(t, c.Start(ctx))

	inspect, err := c.Inspect(ctx)
	assert.NilError(t, err)
	assert.Assert(t, inspect.State.Running)
}
