package container

import (
	"context"
	"testing"

	"github.com/cpuguy83/go-docker"
	"gotest.tools/assert"
)

func TestCreate(t *testing.T) {
	ctx := context.Background()
	client := docker.G(ctx)
	ctx = docker.WithClient(ctx, client)

	c, err := Create(ctx)
	assert.Assert(t, err != nil, err.Error())
	assert.Assert(t, c == nil)

	c, err = Create(ctx, WithCreateImage("busybox:latest"))
	assert.NilError(t, err)
	defer Remove(ctx, c.ID(), WithRemoveForce)
}
