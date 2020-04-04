package httputil

import (
	"encoding/json"
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

// CheckResponseError checks the http response for standard error codes.
//
// For the most part this should return error implemented from the `errdefs` package
func CheckResponseError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return nil
	}

	var e errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
		resp.Body.Close()
		return errors.Wrap(fromStatusCode(err, resp.StatusCode), "error unmarshaling server error response")
	}

	return fromStatusCode(&e, resp.StatusCode)
}

func fromStatusCode(err error, statusCode int) error {
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
