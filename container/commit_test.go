package container

import (
	"context"
	"testing"

	"gotest.tools/assert"
)

func TestCommit(t *testing.T) {
	t.Skip("image api not implemented, don't create stuff that we can't remove or even check")

	ctx := context.Background()
	s := newTestService(t)

	c, err := s.Create(ctx, WithCreateImage("busybox:latest"))
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
