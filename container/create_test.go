package container

import (
	"context"
	"strings"
	"testing"

	"github.com/cpuguy83/go-docker/container/containerapi"
	"github.com/cpuguy83/go-docker/errdefs"
	"github.com/cpuguy83/go-docker/testutils"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	s, ctx := newTestService(t, context.Background())

	c, err := s.Create(ctx, "")
	assert.Check(t, errdefs.IsInvalid(err), err)
	assert.Check(t, c == nil)
	if c != nil {
		if err := s.Remove(ctx, c.ID(), WithRemoveForce); err != nil && !errdefs.IsNotFound(err) {
			t.Error(err)
		}
	}

	name := strings.ToLower(t.Name()) + testutils.GenerateRandomString()
	c, err = s.Create(ctx, "busybox:latest", WithCreateName(name))
	assert.NilError(t, err)
	defer func() {
		assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
	}()

	assert.Assert(t, c.ID() != "")

	inspect, err := c.Inspect(ctx)
	assert.NilError(t, err)
	assert.Equal(t, name, strings.TrimPrefix(inspect.Name, "/"))

	t.Run("port bindings", func(t *testing.T) {
		t.Parallel()

		c, err := s.Create(ctx, "busybox:latest",
			WithCreatePortForwarding("tcp", 80),
			WithCreatePortForwarding("udp", 81),
			WithCreatePortForwarding("udp", 82, containerapi.PortBinding{HostIP: "127.0.0.1"}),
			WithCreateCmd("top"),
		)
		assert.NilError(t, err)
		defer func() {
			assert.Check(t, s.Remove(ctx, c.ID(), WithRemoveForce))
		}()

		assert.Assert(t, c.ID() != "")
		assert.NilError(t, c.Start(ctx))

		inspect, err := c.Inspect(ctx)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(inspect.Config.ExposedPorts["80/tcp"], struct{}{}))

		port80, ok := inspect.NetworkSettings.Ports["80/tcp"]
		assert.Check(t, ok)
		port81, ok := inspect.NetworkSettings.Ports["81/udp"]
		assert.Check(t, ok)

		port82, ok := inspect.NetworkSettings.Ports["82/udp"]
		assert.Check(t, ok)
		assert.Check(t, cmp.Equal(port82[0].HostIP, "127.0.0.1"))

		// Depending on the version of docker (and ipv6 support), there maybe be one
		// or two bindings.
		var hostPort string
		for _, bind := range port80 {
			if bind.HostPort != "" {
				hostPort = bind.HostPort
				break
			}
		}
		assert.Check(t, hostPort != "", "expected a host port binding for 80/tcp")

		hostPort = ""
		for _, bind := range port81 {
			if bind.HostPort != "" {
				hostPort = bind.HostPort
				break
			}
		}
		assert.Check(t, hostPort != "", "expected a host port binding for 81/udp")
	})
}
