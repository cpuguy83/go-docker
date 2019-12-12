package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/cpuguy83/go-docker/transport"

	"github.com/pkg/errors"
)

type Client struct {
	Transport  transport.Doer // TODO: Having a Doer here is possibly weird... at least naming? Not sure.
	APIVersion string
}

func (c *Client) Do(ctx context.Context, method, uri string, opts ...transport.RequestOpt) (*http.Response, error) {
	if c.APIVersion != "" {
		uri = "/v" + c.APIVersion + uri
	}
	resp, err := c.Transport.Do(ctx, method, uri, opts...)
	if err != nil {
		return nil, err
	}
	if err := checkResponseError(resp); err != nil {
		return nil, errors.Wrapf(err, "error performing request: %s %s", method, uri)
	}
	return resp, nil
}

func (c *Client) DoRaw(ctx context.Context, method, uri string, opts ...transport.RequestOpt) (io.ReadWriteCloser, error) {
	if c.APIVersion != "" {
		uri = "/v" + c.APIVersion + uri
	}
	return c.Transport.DoRaw(ctx, method, uri, opts...)
}

type errorResponse struct {
	Message string `json:"message"`
}

func (e errorResponse) Error() string {
	return e.Message
}

func checkResponseError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode <= 400 {
		return nil
	}

	b, err := ioutil.ReadAll(io.LimitReader(resp.Body, 16*1024))
	if err != nil {
		return errors.Wrap(err, "error reading error response body")
	}

	var e errorResponse
	if err := json.Unmarshal(b, &e); err != nil {
		return errors.Wrap(err, "error unmarshaling server error response")
	}

	return fromStatusCode(&e, resp.StatusCode)
}

func WithJSONBody(v interface{}) func(req *http.Request) error {
	return func(req *http.Request) error {
		data, err := json.Marshal(v)
		if err != nil {
			return errors.Wrap(err, "error marshaling json body")
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(data))
		if req.Header == nil {
			req.Header = http.Header{}
		}
		req.Header.Set("Content-Type", "application/json")
		return nil
	}
}

func WithQueryValue(key, value string) func(req *http.Request) error {
	return func(req *http.Request) error {
		q := req.URL.Query()
		if q == nil {
			q = url.Values{}
		}

		q.Set(key, value)
		req.URL.RawQuery = q.Encode()
		return nil
	}
}
