package container

import (
	"context"
	"strings"
	"testing"

	"github.com/cpuguy83/go-docker/testutils"

	"gotest.tools/assert"
)

func TestCreate(t *testing.T) {
	ctx := context.Background()
	tr := testutils.NewDefaultTestTransport(t)
	s := NewService(tr)

	c, err := s.Create(ctx)
	assert.Check(t, err != nil, err)
	assert.Check(t, c == nil)
	if c != nil {
		s.Remove(ctx, c.ID(), WithRemoveForce)
	}

	name := strings.ToLower(t.Name())
	c, err = s.Create(ctx, WithCreateImage("busybox:latest"), WithCreateName(name))
	assert.NilError(t, err)
	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	assert.Assert(t, c.ID() != "")

	inspect, err := c.Inspect(ctx)
	assert.NilError(t, err)
	assert.Equal(t, name, strings.TrimPrefix(inspect.Name, "/"))
}
