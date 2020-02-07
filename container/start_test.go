package container

import (
	"context"
	"testing"

	"github.com/cpuguy83/go-docker/errdefs"
	"gotest.tools/assert"
)

func TestStart(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	c := s.NewContainer(ctx, "notexist")
	err := c.Start(ctx)
	assert.Assert(t, errdefs.IsNotFound(err), err)

	c, err = s.Create(ctx, WithCreateImage("busybox:latest"))
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
