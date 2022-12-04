package image

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"
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

	err = s.Load(ctx, io.NopCloser(bytes.NewReader(bundle)))
	assert.NilError(t, err, "expecting load to succeed")

	buf := bytes.NewBuffer(nil)
	consume := func(ctx context.Context, rdr io.Reader) error {
		_, err := io.Copy(buf, rdr)
		return err
	}

	err = s.Load(ctx, io.NopCloser(bytes.NewReader(bundle)), func(cfg *LoadConfig) error {
		cfg.ConsumeProgress = consume
		return nil
	})
	assert.NilError(t, err, "expecting load to succeed")
	assert.Equal(t, `{"stream":"Loaded image: hello-world:latest\n"}`, strings.TrimSpace(buf.String()))
}
