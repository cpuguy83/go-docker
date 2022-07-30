package container

import (
	"context"
	"strings"
	"testing"

	"gotest.tools/v3/assert/cmp"

	"github.com/cpuguy83/go-docker/errdefs"
	"gotest.tools/v3/assert"
)

func TestInspect(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	_, err := s.Inspect(ctx, "notExist")
	assert.Check(t, errdefs.IsNotFound(err), err)

	name := strings.ToLower(t.Name())
	c, err := s.Create(ctx, "busybox:latest", WithCreateName(name))
	assert.NilError(t, err)
	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	inspect, err := c.Inspect(ctx)
	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(inspect.ID, c.ID()))
	assert.Check(t, cmp.Equal(strings.TrimPrefix(inspect.Name, "/"), name))

	type inspectTo struct {
		ID string
	}
	to := &inspectTo{}
	inspect, err = c.Inspect(ctx, func(o *InspectConfig) {
		o.To = &to
	})
	assert.NilError(t, err)
	assert.Equal(t, to.ID, c.ID())
}
