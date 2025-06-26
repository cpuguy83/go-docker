package image

import (
	"archive/tar"
	"context"
	"io"
	"testing"

	"gotest.tools/v3/assert"
)

func TestExport(t *testing.T) {
	t.Parallel()

	s := newTestService(t)

	ctx := context.Background()
	digest := "faa03e786c97f07ef34423fccceeec2398ec8a5759259f94d99078f264e9d7af"
	err := s.Pull(ctx, Remote{Locator: "hello-world", Host: "docker.io", Tag: "sha256:" + digest})
	assert.NilError(t, err)

	defer s.Remove(ctx, "hello-world@sha256:"+digest)

	rdr, err := s.Export(ctx, WithExportRefs("hello-world@sha256:faa03e786c97f07ef34423fccceeec2398ec8a5759259f94d99078f264e9d7af"))
	assert.NilError(t, err)
	defer rdr.Close()

	tar := tar.NewReader(rdr)

	var found bool
	for {
		hdr, err := tar.Next()
		if err == io.EOF {
			break
		}
		assert.NilError(t, err)

		t.Log(hdr.Name)

		if hdr.Name == "manifest.json" {
			found = true
			break
		}
	}

	assert.Assert(t, found)
}
