package container

import (
	"context"
	"testing"

	"github.com/cpuguy83/go-docker/errdefs"
	"gotest.tools/v3/assert"
)

func TestRemove(t *testing.T) {
	s, ctx := newTestService(t, context.Background())

	err := s.Remove(ctx, "notexist")
	assert.Check(t, errdefs.IsNotFound(err))

	c, err := s.Create(ctx, "busybox:latest")
	assert.NilError(t, err)
	assert.Check(t, s.Remove(ctx, c.ID()), "leaked container: %s", c.ID())

	c, err = s.Create(ctx, "busybox:latest", WithCreateCmd("top"))
	assert.NilError(t, err)
	assert.Assert(t, c.Start(ctx))
	err = s.Remove(ctx, c.ID())
	assert.Assert(t, errdefs.IsConflict(err), err)
	err = s.Remove(ctx, c.ID(), WithRemoveForce)
	assert.NilError(t, err)
}
