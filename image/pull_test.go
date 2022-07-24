package image

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func TestPull(t *testing.T) {
	svc := newTestService(t)

	ctx := context.Background()
	err := svc.Pull(ctx, Remote{Locator: "busybox", Tag: "latest"})
	assert.NilError(t, err)
}
