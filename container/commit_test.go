package container

import (
	"context"
	"slices"
	"testing"

	"github.com/cpuguy83/go-docker/image"
	"github.com/cpuguy83/go-docker/testutils"
	"gotest.tools/v3/assert"
)

func TestCommit(t *testing.T) {
	t.Parallel()

	s, ctx := newTestService(t, context.Background())

	c, err := s.Create(ctx, "busybox:latest")
	assert.NilError(t, err)
	defer s.Remove(ctx, c.ID(), WithRemoveForce)

	repo := "test"
	tag := "commit" + testutils.GenerateRandomString()

	ref, err := c.Commit(ctx, func(cfg *CommitConfig) {
		cfg.Reference = &CommitImageReference{
			Repo: repo,
			Tag:  tag,
		}
	})
	assert.NilError(t, err)

	resp, err := image.NewService(s.tr).Remove(ctx, ref, image.WithRemoveForce)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Contains(resp.Deleted, ref) {
		t.Errorf("expected ref returned by commit (%s) to be in deleted list: %v", ref, resp.Deleted)
	}

	if !slices.Contains(resp.Untagged, repo+":"+tag) {
		t.Errorf("expected tagged image (%s) to be in untagged list: %v", "test:commit", resp.Untagged)
	}
}
