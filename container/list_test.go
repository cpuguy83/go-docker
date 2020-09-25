package container

import (
	"context"
	"fmt"
	"testing"

	"gotest.tools/assert"
)

func TestList(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	c, err := s.Create(ctx, WithCreateImage("busybox:latest"),
		WithCreateCmd("/bin/sh", "-c", "trap 'exit 0' SIGTERM; while true; do sleep 0.1; done"),
	)
	assert.NilError(t, err)

	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	err = c.Start(ctx)
	assert.NilError(t, err)

	containers, err := s.List(ctx)
	found := false
	for _, container := range containers {
		fmt.Printf("\n\nid: %s\n", c.ID())
		fmt.Printf("\n\nlooking for: %s\n", container.ID)
		if container.ID == c.ID() {
			found = true
			break
		}
	}

	assert.Assert(t, found, "expected container to be found but it wasn't")
}

func TestListLimit(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)
	n := 4

	for i := 0; i < n; i++ {
		c, err := s.Create(ctx, WithCreateImage("busybox:latest"),
			WithCreateCmd("/bin/sh", "-c", "trap 'exit 0' SIGTERM; while true; do sleep 0.1; done"),
		)
		assert.NilError(t, err)

		defer func() {
			assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
		}()

		err = c.Start(ctx)
		assert.NilError(t, err)
	}

	containers, err := s.List(ctx, func(config *ListConfig) {
		config.Limit = 2
	})
	assert.NilError(t, err)
	assert.Assert(t, len(containers) == 2, "expected container to be found but it wasn't")
	assert.Equal(t, containers[0].SizeRootFs, 0, "expected container's SizeRootFs to be a zero value")
	assert.Assert(t, containers[0].SizeRw == 0, "expected container's SizeRw to be a positive integer")

}

func TestListSize(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)
	n := 1

	for i := 0; i < n; i++ {
		c, err := s.Create(ctx, WithCreateImage("busybox:latest"),
			WithCreateCmd("/bin/sh", "-c", "trap 'exit 0' SIGTERM; while true; do sleep 0.1; done"),
		)
		assert.NilError(t, err)

		defer func() {
			assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
		}()

		err = c.Start(ctx)
		assert.NilError(t, err)
	}

	containers, err := s.List(ctx, func(config *ListConfig) {
		config.Limit = 1
		config.Size = true
	})
	assert.NilError(t, err)
	assert.Assert(t, len(containers) == 1, "expected container to be found but it wasn't")
	assert.Assert(t, containers[0].SizeRw == 0, "expected container's SizeRw to exist")
	assert.Assert(t, containers[0].SizeRootFs > 0, "expected container's SizeRootFs to be a positive integer")
}

func TestListFilter(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)
	n := 2

	for i := 0; i < n; i++ {
		c, err := s.Create(ctx,
			WithCreateImage("busybox:latest"),
			WithCreateCmd("/bin/sh", "-c", "trap 'exit 0' SIGTERM; while true; do sleep 0.1; done"),
			WithCreateName(fmt.Sprintf("foobar-%d", i)),
		)
		assert.NilError(t, err)

		defer func() {
			assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
		}()

		err = c.Start(ctx)
		assert.NilError(t, err)
	}

	containers, err := s.List(ctx, func(config *ListConfig) {
		config.Filter = ListFilter{Name: []string{"foobar-0"}}
	})
	assert.NilError(t, err)
	assert.Assert(t, len(containers) == 1, "expected container to be %d but received %d", 1, len(containers))
	assert.Assert(t, containers[0].Names[0] == "/foobar-0")
}
