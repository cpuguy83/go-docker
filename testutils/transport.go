package testutils

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/cpuguy83/go-docker/transport"
)

var (
	// regex to match any non-empty identity token
	jsonIdentityTokenRegex = regexp.MustCompile(`"((?i)identitytoken|password|auth)":\ ?".*"`)
)

func NewTransport(t LogT, client transport.Doer) *Transport {
	return &Transport{client, t}
}

type LogT interface {
	Log(...interface{})
	Logf(string, ...interface{})
	Helper()
}

type Transport struct {
	d transport.Doer
	t LogT
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

func (t *Transport) DoRaw(ctx context.Context, method, uri string, opts ...transport.RequestOpt) (io.ReadWriteCloser, error) {
	t.t.Helper()
	opts = append(opts, t.logRequest)
	return t.d.DoRaw(ctx, method, uri, opts...)
}

func (t *Transport) logRequest(req *http.Request) error {
	t.t.Helper()
	t.t.Log(req.Method, req.URL.String())

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
	var status string
	if resp != nil {
		status = resp.Status
	}
	t.t.Log(status, err)

	if resp == nil {
		return resp, err
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		return resp, nil
	}

	buf := bytes.NewBuffer(nil)
	if _, err := ioutil.ReadAll(io.TeeReader(resp.Body, buf)); err != nil {
		return resp, err
	}

	resp.Body = wrapReader(buf, resp.Body.Close)

	t.t.Log(filterBuf(buf).String())

	return resp, err
}

func filterBuf(buf *bytes.Buffer) *bytes.Buffer {
	return bytes.NewBuffer(jsonIdentityTokenRegex.ReplaceAll(buf.Bytes(), []byte(`"${1}": "<REDACTED>"`)))
}
