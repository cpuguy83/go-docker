package transport

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cpuguy83/go-docker/errdefs"
	"github.com/pkg/errors"
)

type errorResponse struct {
	Message string `json:"message"`
}

func (e errorResponse) Error() string {
	return e.Message
}

func checkResponseError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
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

	return FromStatusCode(&e, resp.StatusCode)
}

// FromStatusCode creates an errdef error, based on the provided HTTP status-code
func FromStatusCode(err error, statusCode int) error {
	if err == nil {
		return err
	}
	err = errors.Wrapf(err, "error in response, status code: %d", statusCode)
	switch statusCode {
	case http.StatusNotFound:
		err = errdefs.AsNotFound(err)
	case http.StatusBadRequest:
		err = errdefs.AsInvalidInput(err)
	case http.StatusConflict:
		err = errdefs.AsConflict(err)
	case http.StatusUnauthorized:
		err = errdefs.AsUnauthorized(err)
	case http.StatusServiceUnavailable:
		err = errdefs.AsUnavailable(err)
	case http.StatusForbidden:
		err = errdefs.AsForbidden(err)
	case http.StatusNotModified:
		err = errdefs.AsNotModified(err)
	case http.StatusNotImplemented:
		err = errdefs.AsNotImplemented(err)
	default:
		if statusCode >= 400 && statusCode < 500 {
			err = errdefs.AsInvalidInput(err)
		}
	}
	return err
}
