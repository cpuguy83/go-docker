package image

import (
	"context"
	"fmt"
	"testing"

	"github.com/opencontainers/go-digest"
	"gotest.tools/v3/assert"
)

func TestPull(t *testing.T) {
	svc := newTestService(t)

	ctx := context.Background()
	var dgst string
	digestFn := func(ctx context.Context, s string) error {
		if _, err := digest.Parse(s); err != nil {
			return fmt.Errorf("%w: %s", err, s)
		}
		dgst = s
		return nil
	}
	err := svc.Pull(ctx, Remote{Locator: "busybox", Tag: "latest"}, WithPullProgressMessage(PullProgressDigest(digestFn)))
	assert.NilError(t, err)
	assert.Assert(t, dgst != "")

	// This is used by some other tests right now, so don't remove it
	/// _, err = svc.Remove(ctx, "busybox:latest")
	assert.NilError(t, err)
}
