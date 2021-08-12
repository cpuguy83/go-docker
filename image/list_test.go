package image_test

import (
	"context"
	"testing"

	"github.com/cpuguy83/go-docker/image"
	"gotest.tools/assert"
)

func TestList(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	// First create a few images that we can list later.
	err := s.Create(ctx, func(config *image.CreateConfig) {
		config.FromImage = "busybox"
		config.Tag = "latest"
		config.Repo = "dockerhub.io"
	})
	assert.NilError(t, err, "expected pulling busybox to succeed")
	err = s.Create(ctx, func(config *image.CreateConfig) {
		config.FromImage = "hello-world"
		config.Tag = "latest"
		config.Repo = "dockerhub.io"
	})
	assert.NilError(t, err, "expected pulling hello-world to succeed")

	images, err := s.List(ctx, func(config *image.ListConfig) {
		config.Filter.Reference = append(config.Filter.Reference, "busybox:latest", "hello-world:latest")
	})
	assert.NilError(t, err, "expected listing images with no options to succeed")
	assert.Assert(t, len(images) == 2, "expected created images to be listed")
}
