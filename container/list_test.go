package container

import (
	"context"
	"strings"
	"testing"

	"gotest.tools/assert"

	"github.com/cpuguy83/go-docker/common/filters"
)

func TestList(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	name := strings.ToLower(t.Name())
	c, err := s.Create(ctx, WithCreateImage("busybox:latest"), WithCreateName(name))
	assert.NilError(t, err)
	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	containers, err := s.List(ctx, WithListAll)
	assert.NilError(t, err)
	assert.Assert(t, len(containers) != 0)

	found := false
	for _, ct := range containers {
		if ct.ID == c.ID() {
			found = true
		}
	}

	assert.Assert(t, found)
}

func TestListLabelFilter(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	name := strings.ToLower(t.Name())
	c, err := s.Create(ctx, WithCreateImage("busybox:latest"), WithCreateName(name), WithCreateLabels(map[string]string{"test-list": "foo"}))
	assert.NilError(t, err)
	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	containers, err := s.List(ctx, WithListAll, WithListFilters(filters.NewArgs(filters.Arg("label", "test-list=foo"))))
	assert.NilError(t, err)
	assert.Assert(t, len(containers) != 0)

	found := false
	for _, ct := range containers {
		if ct.ID == c.ID() {
			found = true
		}
	}

	assert.Assert(t, found)
}

func TestListLabelLimit(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	for i := 0; i < 2; i++ {
		c, err := s.Create(ctx, WithCreateImage("busybox:latest"), WithCreateLabels(map[string]string{"test-list": "foo"}))
		assert.NilError(t, err)
		defer func() {
			assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
		}()
	}

	containers, err := s.List(ctx, WithListAll, WithListLimit(1))
	assert.NilError(t, err)
	assert.Equal(t, 1, len(containers))

	assert.Equal(t, "foo", containers[0].Labels["test-list"])
}
