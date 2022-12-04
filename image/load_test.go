package image

import (
	"bytes"
	"context"
	"io"
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

	rdr, err := s.Export(ctx, WithExportRefs("hello-world:latest"))
	assert.NilError(t, err)

	buf := bytes.NewBuffer(nil)
	consume := func(ctx context.Context, rdr io.Reader) error {
		_, err := io.Copy(buf, rdr)
		return err
	}

	err = s.Load(ctx, rdr, func(cfg *LoadConfig) error {
		cfg.ConsumeProgress = consume
		return nil
	})
	defer s.Remove(ctx, "hello-world:latest")
	assert.NilError(t, err, "expecting load to succeed")
	assert.Equal(t, `{"stream":"Loaded image: hello-world:latest\n"}`, strings.TrimSpace(buf.String()))
}
