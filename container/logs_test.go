package container

import (
	"bufio"
	"context"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

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

	logs, mux, err := c.Logs(ctx, func(config *LogReadConfig) {
		config.ShowStdout = true
	})
	assert.NilError(t, err)
	assert.Check(t, mux)

	buf, err := ioutil.ReadAll(logs)
	assert.NilError(t, err)

	assert.Assert(t, strings.Contains(string(buf), "hello there"), "expected container logs to contain 'hello there'")
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

	logs, mux, err := c.Logs(ctx, func(config *LogReadConfig) {
		config.ShowStderr = true
	})
	assert.NilError(t, err)
	assert.Check(t, mux)

	buf, err := ioutil.ReadAll(logs)
	assert.NilError(t, err)

	assert.Assert(t, strings.Contains(string(buf), "bad things"), "expected container logs to contain 'bad things'")
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

	logs, mux, err := c.Logs(ctx, func(config *LogReadConfig) {
		config.ShowStdout = true
	})
	assert.NilError(t, err)
	assert.Check(t, mux)

	buf, err := ioutil.ReadAll(logs)
	assert.NilError(t, err)

	assert.Assert(t, !strings.Contains(string(buf), "bad things"), "expected container logs to not contain 'bad things'")
}

func TestLogsSince(t *testing.T) {
	s, ctx := newTestService(t, context.Background())

	c, err := s.Create(ctx, "busybox:latest",
		WithCreateCmd("/bin/sh", "-c", "echo 'hello there'; sleep 2; echo 'why hello'"),
	)
	assert.NilError(t, err)

	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	err = c.Start(ctx)
	assert.NilError(t, err)

	time.Sleep(2 * time.Second)
	ts := time.Now().Unix()

	logs, mux, err := c.Logs(ctx, func(config *LogReadConfig) {
		config.ShowStdout = true
		config.Since = strconv.FormatInt(ts, 10)
	})
	assert.NilError(t, err)
	assert.Check(t, mux)

	buf, err := ioutil.ReadAll(logs)
	assert.NilError(t, err)

	assert.Assert(t, !strings.Contains(string(buf), "hello there"), "expected container logs to not contain 'hello there'")
	assert.Assert(t, strings.Contains(string(buf), "why hello"), "expected container logs to contain 'why hello'")
}

func TestLogsUntil(t *testing.T) {
	s, ctx := newTestService(t, context.Background())

	c, err := s.Create(ctx, "busybox:latest",
		WithCreateCmd("/bin/sh", "-c", "echo 'hello there'; sleep 2; echo 'why hello'"),
	)
	assert.NilError(t, err)

	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	err = c.Start(ctx)
	assert.NilError(t, err)

	time.Sleep(1 * time.Second)
	ts := time.Now().Unix()
	time.Sleep(1 * time.Second)

	logs, mux, err := c.Logs(ctx, func(config *LogReadConfig) {
		config.ShowStdout = true
		config.Until = strconv.FormatInt(ts, 10)
	})
	assert.NilError(t, err)
	assert.Check(t, mux)

	buf, err := ioutil.ReadAll(logs)
	assert.NilError(t, err)

	assert.Assert(t, strings.Contains(string(buf), "hello there"), "expected container logs to contain 'hello there'")
	assert.Assert(t, !strings.Contains(string(buf), "why hello"), "expected container logs to not contain 'why hello'")
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

	logs, mux, err := c.Logs(ctx, func(config *LogReadConfig) {
		config.ShowStdout = true
		config.Timestamps = true
	})
	assert.NilError(t, err)
	assert.Check(t, mux)

	r := bufio.NewReader(logs)

	header := make([]byte, 8)
	_, err = io.ReadFull(r, header)
	assert.NilError(t, err)

	ts, err := r.ReadString(' ')
	assert.NilError(t, err)

	parsedTime, err := time.Parse("2006-01-02T15:04:05.999999999Z", ts[:len(ts)-1])
	assert.NilError(t, err)

	now := time.Now().UTC()
	t.Logf("%s", now)
	assert.Assert(t, parsedTime.Year() == now.Year(), "expected parsed year to be %d but received %d", now.Year(), parsedTime.Year())
	assert.Assert(t, parsedTime.Month() == now.Month(), "expected parsed month to be %s but received %s", now.Month(), parsedTime.Month())
	assert.Assert(t, parsedTime.Day() == now.Day(), "expected parsed day to be %d but received %d", now.Day(), parsedTime.Day())
}

func TestLogsTTYMux(t *testing.T) {
	s, ctx := newTestService(t, context.Background())

	c, err := s.Create(ctx, "busybox:latest",
		WithCreateCmd("/bin/sh", "-c", "echo 'hello there'"),
		WithCreateTTY,
	)
	assert.NilError(t, err)

	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	err = c.Start(ctx)
	assert.NilError(t, err)

	logs, mux, err := c.Logs(ctx, func(config *LogReadConfig) {
		config.ShowStdout = true
	})
	assert.NilError(t, err)
	defer logs.Close()
	assert.Check(t, !mux)
}
