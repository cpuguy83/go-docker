package docker

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/docker/docker/api/types"
	"gotest.tools/assert"
)

func testDoer struct {}

func TestPing(t *testing.T) {
	statuses := []int{http.StatusOK, http.StatusConflict}
	for _, s := range statuses {
		t.Run("Status"+strconv.Itoa(s), func(t *testing.T) {
			c := &Client{
				Client: &http.Client{
					Transport: &testRoundTripper{
						handler: func(req *http.Request) *http.Response {
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
						},
					},
				},
			}

			p, err := c.Ping(context.Background())
			if s == http.StatusOK {
				assert.NilError(t, err)
			} else {
				assert.Assert(t, err != nil)
			}

			assert.Equal(t, p.OSType, "the best (one for the job)!")
			assert.Equal(t, p.APIVersion, "banana")
			assert.Equal(t, p.BuilderVersion, types.BuilderVersion("apple"))
			assert.Equal(t, p.Experimental, true)
		})
	}
}
