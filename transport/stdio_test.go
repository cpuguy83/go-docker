package transport

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestFromDockerCLI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello " + req.URL.Path))
	}))
	defer srv.Close()

	u, err := url.Parse(srv.URL)
	assert.NilError(t, err)

	errBuf := bytes.NewBuffer(nil)
	tr := FromDockerCLI(func(dcc *DockerCLIConnectionConfig) error {
		dcc.Env = []string{"DOCKER_HOST=" + "tcp://" + u.Host}
		dcc.StderrPipe = errBuf
		return nil
	})
	t.Cleanup(func() {
		if t.Failed() && errBuf.Len() > 0 {
			t.Log(errBuf.String())
		}
	})

	ctx := context.Background()

	ctxT, cancel := context.WithTimeout(ctx, 10*time.Second)
	resp, err := tr.Do(ctxT, "GET", "/foo")
	cancel()
	assert.NilError(t, err)
	defer resp.Body.Close()

	data := "hello /foo"
	buf := make([]byte, len(data))
	_, err = io.ReadFull(resp.Body, buf)
	assert.NilError(t, err)
	assert.Equal(t, string(buf), data, &readerStringer{resp.Body})
}

type readerStringer struct {
	io.Reader
}

func (r *readerStringer) String() string {
	out, _ := io.ReadAll(r.Reader)
	return string(out)
}
