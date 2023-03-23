package testutils

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"

	"github.com/cpuguy83/go-docker/transport"
)

var (
	// regex to match any non-empty identity token
	jsonIdentityTokenRegex = regexp.MustCompile(`"((?i)identitytoken|password|auth)":\ ?".*"`)
)

func NewTransport(t LogT, client transport.Doer, noTap bool) *Transport {
	return &Transport{client, t, noTap}
}

type LogT interface {
	Log(...interface{})
	Logf(string, ...interface{})
	Helper()
}

type Transport struct {
	d     transport.Doer
	t     LogT
	noTap bool
}

type readCloserWrapper struct {
	io.Reader
	close func() error
}

func (w *readCloserWrapper) Close() error {
	if w.close != nil {
		return w.close()
	}
	return nil
}

func wrapReader(r io.Reader, f func() error) io.ReadCloser {
	return &readCloserWrapper{r, f}
}

func (t *Transport) Do(ctx context.Context, method, uri string, opts ...transport.RequestOpt) (*http.Response, error) {
	t.t.Helper()
	opts = append(opts, t.logRequest)
	return t.logResponse(t.d.Do(ctx, method, uri, opts...))
}

func (t *Transport) DoRaw(ctx context.Context, method, uri string, opts ...transport.RequestOpt) (net.Conn, error) {
	t.t.Helper()
	opts = append(opts, t.logRequest)
	conn, err := t.d.DoRaw(ctx, method, uri, opts...)
	if err != nil {
		return conn, err
	}

	if t.noTap {
		return conn, nil
	}

	p1, p2 := net.Pipe()

	go func() {
		io.Copy(p2, io.TeeReader(conn, &testingWriter{t.t}))
		p2.Close()
	}()

	go func() {
		io.Copy(conn, io.TeeReader(p2, &testingWriter{t.t}))
		conn.Close()
	}()

	return p1, nil
}

type testingWriter struct {
	t LogT
}

func (t *testingWriter) Write(p []byte) (int, error) {
	t.t.Helper()
	t.t.Log(string(p))
	return len(p), nil
}

func (t *Transport) logRequest(req *http.Request) error {
	t.t.Helper()
	t.t.Log(req.Method, req.URL.String())
	t.t.Log(req.Header)

	if req.Header.Get("Content-Type") != "application/json" {
		return nil
	}

	buf := bytes.NewBuffer(nil)
	if _, err := ioutil.ReadAll(io.TeeReader(req.Body, buf)); err != nil {
		return err
	}

	req.Body = wrapReader(buf, req.Body.Close)

	t.t.Log(filterBuf(buf).String())
	return nil
}

func (t *Transport) logResponse(resp *http.Response, err error) (*http.Response, error) {
	t.t.Helper()

	if resp == nil {
		return resp, err
	}

	if resp != nil {
		t.t.Log(resp.Status, err)
		t.t.Log(resp.Header)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		return resp, nil
	}

	buf := bytes.NewBuffer(nil)
	b := resp.Body
	rdr := io.TeeReader(b, buf)
	resp.Body = wrapReader(rdr, func() error {
		t.t.Log(filterBuf(buf).String())
		return b.Close()
	})

	return resp, err
}

func filterBuf(buf *bytes.Buffer) *bytes.Buffer {
	return bytes.NewBuffer(jsonIdentityTokenRegex.ReplaceAll(buf.Bytes(), []byte(`"${1}": "<REDACTED>"`)))
}
