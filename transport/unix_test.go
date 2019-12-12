package transport

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/assert"
)

func TestUnixTransport(t *testing.T) {
	dir, err := ioutil.TempDir("", t.Name())
	assert.NilError(t, err)
	defer os.RemoveAll(dir)

	sockPath := filepath.Join(dir, "test.sock")
	l, err := net.Listen("unix", sockPath)
	assert.NilError(t, err)
	defer l.Close()

	go http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello " + req.URL.Path))
	}))

	ctx := context.Background()

	tr, err := UnixSocketTransport(sockPath)
	assert.NilError(t, err)

	resp, err := tr.Do(ctx, "GET", "/foo")
	assert.NilError(t, err)
	defer resp.Body.Close()

	data := "hello /foo"
	buf := make([]byte, len(data))
	_, err = io.ReadFull(resp.Body, buf)
	assert.NilError(t, err)
	assert.Equal(t, string(buf), data)
}
