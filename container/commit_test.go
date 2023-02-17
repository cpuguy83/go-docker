package container

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func TestCommit(t *testing.T) {
	t.Skip("image api not implemented, don't create stuff that we can't remove or even check")

	s, ctx := newTestService(t, context.Background())

	c, err := s.Create(ctx, "busybox:latest")
	assert.NilError(t, err)

	// TODO: check image reference?
	_, err = c.Commit(ctx, func(cfg *CommitConfig) {
		cfg.Reference = &CommitImageReference{
			Repo: "test",
			Tag:  "commit",
		}
	})
	defer s.Remove(ctx, c.ID(), WithRemoveForce)
	assert.NilError(t, err)
}
