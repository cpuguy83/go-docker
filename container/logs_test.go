package container

import (
	"bufio"
	"context"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func waitForContainerExit(ctx context.Context, t *testing.T, c *Container) {
	t.Helper()

	wait, err := c.Wait(ctx, WithWaitCondition(WaitConditionNotRunning))
	assert.NilError(t, err)
	_, err = wait.ExitCode()
	assert.NilError(t, err)
}

func TestStdoutLogs(t *testing.T) {
	s, ctx := newTestService(t, context.Background())

	c, err := s.Create(ctx, "busybox:latest",
		WithCreateCmd("/bin/sh", "-c", "echo 'hello there'"),
	)
	assert.NilError(t, err)

	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	err = c.Start(ctx)
	assert.NilError(t, err)

	waitForContainerExit(ctx, t, c)

	r, w := io.Pipe()
	err = c.Logs(ctx, func(config *LogReadConfig) {
		config.Stdout = w
	})
	assert.NilError(t, err)
	defer r.Close()

	data, err := io.ReadAll(r)
	assert.NilError(t, err)

	assert.Assert(t, cmp.Contains(string(data), "hello there"), "expected container logs to contain 'hello there'")
}

func TestStderrLogs(t *testing.T) {
	s, ctx := newTestService(t, context.Background())

	c, err := s.Create(ctx, "busybox:latest",
		WithCreateCmd("/bin/sh", "-c", ">&2 echo 'bad things'"),
	)
	assert.NilError(t, err)

	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	err = c.Start(ctx)
	assert.NilError(t, err)

	waitForContainerExit(ctx, t, c)

	r, w := io.Pipe()
	err = c.Logs(ctx, func(config *LogReadConfig) {
		config.Stderr = w
	})
	defer r.Close()
	assert.NilError(t, err)

	data, err := io.ReadAll(r)
	assert.NilError(t, err)

	assert.Assert(t, cmp.Contains(string(data), "bad things"), "expected container logs to contain 'bad things'")
}

func TestStdoutStderrLogs(t *testing.T) {
	s, ctx := newTestService(t, context.Background())

	c, err := s.Create(ctx, "busybox:latest",
		WithCreateCmd("/bin/sh", "-c", ">&2 echo 'bad things'"),
	)
	assert.NilError(t, err)

	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	err = c.Start(ctx)
	assert.NilError(t, err)

	r, w := io.Pipe()
	err = c.Logs(ctx, func(config *LogReadConfig) {
		config.Stdout = w
	})
	defer r.Close()
	assert.NilError(t, err)

	data, err := io.ReadAll(r)
	assert.NilError(t, err)

	assert.Equal(t, string(data), "")
}

func TestLogsSince(t *testing.T) {
	s, ctx := newTestService(t, context.Background())

	c, err := s.Create(ctx, "busybox:latest",
		WithCreateCmd("/bin/sh", "-c", "echo 'hello there'; sleep 2; echo 'why hello'"),
	)
	assert.NilError(t, err)

	wait, err := c.Wait(ctx, WithWaitCondition(WaitConditionNextExit))
	assert.NilError(t, err)

	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	err = c.Start(ctx)
	assert.NilError(t, err)

	inspect, err := c.Inspect(ctx)
	assert.NilError(t, err)
	started, err := time.Parse(time.RFC3339Nano, inspect.State.StartedAt)
	assert.NilError(t, err)
	ts := started.Add(2 * time.Second).Unix()

	_, err = wait.ExitCode()
	assert.NilError(t, err)

	r, w := io.Pipe()
	defer r.Close()
	err = c.Logs(ctx, func(config *LogReadConfig) {
		config.Stdout = w
		config.Since = strconv.FormatInt(ts, 10)
	})
	assert.NilError(t, err)

	data, err := io.ReadAll(r)
	assert.NilError(t, err)

	assert.Assert(t, !strings.Contains(string(data), "hello there"), "expected container logs to not contain 'hello there'")
	assert.Assert(t, strings.Contains(string(data), "why hello"), "expected container logs to contain 'why hello'")
}

func TestLogsUntil(t *testing.T) {
	s, ctx := newTestService(t, context.Background())

	c, err := s.Create(ctx, "busybox:latest",
		WithCreateCmd("/bin/sh", "-c", "echo 'hello there'; sleep 2; echo 'why hello'"),
	)
	assert.NilError(t, err)

	wait, err := c.Wait(ctx, WithWaitCondition(WaitConditionNextExit))
	assert.NilError(t, err)

	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	err = c.Start(ctx)
	assert.NilError(t, err)

	inspect, err := c.Inspect(ctx)
	assert.NilError(t, err)
	started, err := time.Parse(time.RFC3339Nano, inspect.State.StartedAt)
	assert.NilError(t, err)

	_, err = wait.ExitCode()
	assert.NilError(t, err)

	r, w := io.Pipe()
	defer r.Close()
	err = c.Logs(ctx, func(config *LogReadConfig) {
		config.Stdout = w
		config.Until = strconv.FormatInt(started.Add(time.Second).Unix(), 10)
	})
	assert.NilError(t, err)

	data, err := io.ReadAll(r)
	assert.NilError(t, err)

	assert.Assert(t, strings.Contains(string(data), "hello there"), "expected container logs to contain 'hello there'")
	assert.Assert(t, !strings.Contains(string(data), "why hello"), "expected container logs to not contain 'why hello'")
}

func TestLogsTimestamps(t *testing.T) {
	s, ctx := newTestService(t, context.Background())

	c, err := s.Create(ctx, "busybox:latest",
		WithCreateCmd("/bin/sh", "-c", "echo 'hello there'"),
	)
	assert.NilError(t, err)

	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	err = c.Start(ctx)
	assert.NilError(t, err)

	waitForContainerExit(ctx, t, c)

	pr, pw := io.Pipe()
	defer pr.Close()
	err = c.Logs(ctx, func(config *LogReadConfig) {
		config.Stdout = pw
		config.Timestamps = true
	})
	assert.NilError(t, err)

	r := bufio.NewReader(pr)

	ts, err := r.ReadString(' ')
	assert.NilError(t, err)

	parsedTime, err := time.Parse("2006-01-02T15:04:05.999999999Z", ts[:len(ts)-1])
	assert.NilError(t, err)

	now := time.Now().UTC()
	assert.Assert(t, parsedTime.Year() == now.Year(), "expected parsed year to be %d but received %d", now.Year(), parsedTime.Year())
	assert.Assert(t, parsedTime.Month() == now.Month(), "expected parsed month to be %s but received %s", now.Month(), parsedTime.Month())
	assert.Assert(t, parsedTime.Day() == now.Day(), "expected parsed day to be %d but received %d", now.Day(), parsedTime.Day())
}
