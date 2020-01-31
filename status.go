package docker

import (
	"net/http"

	"github.com/cpuguy83/go-docker/errdefs"
	"github.com/pkg/errors"
)

// fromStatusCode creates an errdef error, based on the provided HTTP status-code
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
