package image

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func TestList(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	// First create a few images that we can list later.
	err := s.Pull(ctx, Remote{
		Locator: "busybox",
		Host:    "docker.io",
		Tag:     "latest",
	})
	assert.NilError(t, err, "expected pulling busybox to succeed")
	err = s.Pull(ctx, Remote{
		Locator: "hello-world",
		Host:    "docker.io",
		Tag:     "latest",
	})
	assert.NilError(t, err, "expected pulling hello-world to succeed")

	images, err := s.List(ctx, func(config *ListConfig) {
		config.Filter.Reference = append(config.Filter.Reference, "busybox:latest", "hello-world:latest")
	})
	assert.NilError(t, err, "expected listing images with no options to succeed")
	assert.Assert(t, len(images) == 2, "expected created images to be listed")
}
