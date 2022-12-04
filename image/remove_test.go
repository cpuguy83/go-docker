package image

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func TestRemove(t *testing.T) {
	s := newTestService(t)

	ctx := context.Background()
	err := s.Pull(ctx, Remote{Locator: "hello-world", Host: "docker.io", Tag: "latest"})
	assert.NilError(t, err)

	rm, err := s.Remove(ctx, "hello-world:latest")
	assert.NilError(t, err)
	assert.Assert(t, len(rm.Deleted) > 0 || len(rm.Untagged) > 0)
}
