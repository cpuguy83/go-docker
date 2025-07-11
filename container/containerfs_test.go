package container

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/cpuguy83/go-docker/errdefs"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestContainerFS(t *testing.T) {
	s, ctx := newTestService(t, context.Background())

	c, err := s.Create(ctx, "busybox:latest")
	assert.NilError(t, err)
	defer s.Remove(ctx, c.ID(), WithRemoveForce)

	err = c.Stat(ctx, "/hello.txt", nil)
	assert.Check(t, errdefs.IsNotFound(err))

	fileDt := []byte("Hello, World!\n")
	buf := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buf)

	err = tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "hello.txt",
		Size:     int64(len(fileDt)),
	})
	assert.NilError(t, err)

	_, err = tw.Write(fileDt)
	assert.NilError(t, err)

	assert.NilError(t, tw.Flush())
	assert.NilError(t, tw.Close())

	err = c.Upload(ctx, "/", buf)
	assert.NilError(t, err)

	content, err := c.Download(ctx, "/hello.txt")
	assert.NilError(t, err)
	defer content.Close()

	tr := tar.NewReader(content)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			assert.NilError(t, err)
		}

		assert.Check(t, cmp.Equal(hdr.Name, "hello.txt"))
		assert.Check(t, cmp.Equal(int64(len(fileDt)), hdr.Size))

		dt, err := io.ReadAll(tr)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(string(dt), string(fileDt)))
	}

	// Just make sure we can stat another directory that should
	// aready exist.
	stat := &Stat{}
	err = c.Stat(ctx, "/usr", stat)
	assert.NilError(t, err)
	assert.Check(t, stat.Mode.IsDir())
}
