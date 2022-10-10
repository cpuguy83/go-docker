package image

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"testing"

	"gotest.tools/v3/assert"
)

func TestLoad(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	// Pull a dummy image that we can export.
	err := s.Pull(ctx, Remote{
		Locator: "hello-world",
		Host:    "docker.io",
		Tag:     "latest",
	})
	assert.NilError(t, err, "expected pulling hello-world to succeed")

	cmd := exec.Command("docker", "image", "save", "hello-world:latest")
	bundle, err := cmd.CombinedOutput()
	assert.NilError(t, err, "expected exporting hello-world to succeed")

	err = s.Load(ctx, io.NopCloser(bytes.NewReader(bundle)), func(config *LoadConfig) {
		config.Quiet = true
	})
	assert.NilError(t, err, "expecting load to succeed")
}
