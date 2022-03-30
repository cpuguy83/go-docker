package image_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/cpuguy83/go-docker/image"
	"gotest.tools/assert"
)

func TestLoad(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	// Create a dummy image that we can export.
	err := s.Create(ctx, func(config *image.CreateConfig) {
		config.FromImage = "hello-world"
		config.Tag = "latest"
		config.Repo = "dockerhub.io"
	})
	assert.NilError(t, err, "expected pulling hello-world to succeed")

	bundle, err := s.ExportBundle(ctx, func(config *image.ExportBundleConfig) {
		config.Names = []string{"hello-world:latest", "hello-world:latest"}
	})
	assert.NilError(t, err, "expected exporting hello-world to succeed")

	err = s.Load(ctx, io.NopCloser(bytes.NewReader(bundle)), func(config *image.LoadConfig) {
		config.Quiet = true
	})
	assert.NilError(t, err, "expecting load to succeed")
}
