// +build windows

package transport

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/Microsoft/go-winio"
	"gotest.tools/v3/assert"
)

var testPipeName = `\\.\pipe\winiotestpipe`

func TestWindowsTransport(t *testing.T) {
	l, err := winio.ListenPipe(testPipeName, nil)
	assert.NilError(t, err)
	defer l.Close()

	go http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello " + req.URL.Path))
	}))

	ctx := context.Background()

	tr, err := NpipeTransport(testPipeName)
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
