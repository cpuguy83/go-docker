package system

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"testing"

	"github.com/cpuguy83/go-docker/transport"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

type mockDoer struct {
	doHandlers map[string]func(*http.Request) *http.Response
}

func (m *mockDoer) registerHandler(method, uri string, h func(*http.Request) *http.Response) {
	if m.doHandlers == nil {
		m.doHandlers = make(map[string]func(*http.Request) *http.Response)
	}
	m.doHandlers[path.Join(method, uri)] = h
}

func (m *mockDoer) Do(ctx context.Context, method string, uri string, opts ...transport.RequestOpt) (*http.Response, error) {
	var req http.Request
	for _, o := range opts {
		if err := o(&req); err != nil {
			return nil, err
		}
	}

	h, ok := m.doHandlers[path.Join(method, uri)]
	if !ok {
		return &http.Response{StatusCode: http.StatusNotFound, Status: "not found"}, nil
	}
	return h(&req), nil
}

func (m *mockDoer) DoRaw(ctx context.Context, method string, uri string, opts ...transport.RequestOpt) (io.ReadWriteCloser, error) {
	return nil, errors.New("not supported")
}

func TestPing(t *testing.T) {
	statuses := []int{http.StatusOK, http.StatusConflict}

	for _, s := range statuses {
		t.Run("Status"+strconv.Itoa(s), func(t *testing.T) {
			pingHandler := func(req *http.Request) *http.Response {
				resp := &http.Response{
					StatusCode: s,
					Header:     http.Header{},
					Body:       ioutil.NopCloser(bytes.NewBuffer(nil)),
				}

				resp.Header.Add("OSType", "the best (one for the job)!")
				resp.Header.Add("API-Version", "banana")
				resp.Header.Add("Builder-Version", "apple")
				resp.Header.Add("Docker-Experimental", "true")
				return resp
			}

			tr := &mockDoer{}
			tr.registerHandler(http.MethodGet, "/_ping", pingHandler)
			system := &Service{
				tr: tr,
			}
			p, _ := system.Ping(context.Background())

			assert.Check(t, cmp.Equal(p.OSType, "the best (one for the job)!"))
			assert.Check(t, cmp.Equal(p.APIVersion, "banana"))
			assert.Check(t, cmp.Equal(p.BuilderVersion, "apple"))
			assert.Check(t, cmp.Equal(p.Experimental, true))
		})
	}
}
