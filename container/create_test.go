package container

import (
	"context"
	"testing"

	"github.com/cpuguy83/go-docker/testutils"

	"github.com/cpuguy83/go-docker"
	"gotest.tools/assert"
)

func TestCreate(t *testing.T) {
	ctx := context.Background()
	client := docker.G(ctx)
	client.Transport = testutils.NewTransport(t, client.Transport)
	ctx = docker.WithClient(ctx, client)

	c, err := Create(ctx)
	assert.Check(t, err != nil, err)
	assert.Check(t, c == nil)
	if c != nil {
		Remove(ctx, c.ID(), WithRemoveForce)
	}

	c, err = Create(ctx, WithCreateImage("busybox:latest"))
	assert.NilError(t, err)
	defer Remove(ctx, c.ID(), WithRemoveForce)

	assert.Assert(t, c.ID() != "")
}
