package container

import (
	"context"
	"testing"
	"time"

	"github.com/cpuguy83/go-docker/testutils"

	"github.com/cpuguy83/go-docker/errdefs"
	"gotest.tools/v3/assert"
)

func TestWait(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	c := s.NewContainer(ctx, "notexist")
	_, err := c.Wait(ctx)
	assert.Assert(t, errdefs.IsNotFound(err), err)

	c, err = s.Create(ctx, "busybox:latest",
		WithCreateCmd("/bin/sh", "-c", "trap 'exit 1' EXIT; while true; do sleep 0.1; done"),
	)
	assert.NilError(t, err)

	defer func() {
		ch := make(chan func(t *testing.T), 1)
		go func() {
			_, err := c.Wait(ctx, WithWaitCondition(WaitConditionNextExit))
			ch <- func(t *testing.T) {
				assert.NilError(t, err)

				_, err = c.Inspect(ctx)
				assert.Check(t, errdefs.IsNotFound(err))
			}
		}()
		assert.Check(t, s.Remove(ctx, c.id, WithRemoveForce))
		testutils.Deadline(t, 30*time.Second, ch)
	}()

	es, err := c.Wait(ctx, WithWaitCondition(WaitConditionNotRunning))
	assert.NilError(t, err)
	assert.Equal(t, es.ExitCode(), 0)

	ch := make(chan func(t *testing.T), 1)
	go func() {
		es, err := c.Wait(ctx, WithWaitCondition(WaitConditionNextExit))
		ch <- func(t *testing.T) {
			assert.NilError(t, err)

			inspect, err := c.Inspect(ctx)
			assert.NilError(t, err)
			assert.Equal(t, es.ExitCode(), inspect.State.ExitCode)
		}
	}()

	assert.NilError(t, c.Start(ctx))
	assert.NilError(t, c.Kill(ctx))

	testutils.Deadline(t, 10*time.Second, ch)
}
