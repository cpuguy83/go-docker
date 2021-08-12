package image_test

import (
	"context"
	"testing"

	"github.com/cpuguy83/go-docker/image"
	"gotest.tools/assert"
)

func TestCreate(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	err := s.Create(ctx)
	assert.Assert(t, err != nil, "expected create with no options to fail")

	err = s.Create(ctx, func(config *image.CreateConfig) {
		config.FromImage = "busybox"
		config.Tag = "latest"
		config.Repo = "dockerhub.io"
	})
	assert.NilError(t, err, "expected pulling busybox to succeed")
}
