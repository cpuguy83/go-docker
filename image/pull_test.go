package image

import (
	"context"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestPull(t *testing.T) {
	svc := newTestService(t)

	ctx := context.Background()
	var digest string
	progress := func(ctx context.Context, msg PullProgressMessage) error {
		_, right, ok := strings.Cut(msg.Status, "Digest:")
		if ok {
			digest = right
		}
		return nil
	}
	err := svc.Pull(ctx, Remote{Locator: "busybox", Tag: "latest"}, WithPullProgressMessage(progress))
	assert.NilError(t, err)
	assert.Assert(t, digest != "")
}
