package docker

import (
	"net/http"

	"github.com/docker/docker/errdefs"
	"github.com/pkg/errors"
)

// fromStatusCode creates an errdef error, based on the provided HTTP status-code
//
// TODO: Do not import errdefs from docker/docker
func fromStatusCode(err error, statusCode int) error {
	if err == nil {
		return err
	}
	err = errors.Wrapf(err, "error in response, status code: %d", statusCode)
	switch statusCode {
	case http.StatusNotFound:
		err = errdefs.NotFound(err)
	case http.StatusBadRequest:
		err = errdefs.InvalidParameter(err)
	case http.StatusConflict:
		err = errdefs.Conflict(err)
	case http.StatusUnauthorized:
		err = errdefs.Unauthorized(err)
	case http.StatusServiceUnavailable:
		err = errdefs.Unavailable(err)
	case http.StatusForbidden:
		err = errdefs.Forbidden(err)
	case http.StatusNotModified:
		err = errdefs.NotModified(err)
	case http.StatusNotImplemented:
		err = errdefs.NotImplemented(err)
	case http.StatusInternalServerError:
		if !errdefs.IsSystem(err) && !errdefs.IsUnknown(err) && !errdefs.IsDataLoss(err) && !errdefs.IsDeadline(err) && !errdefs.IsCancelled(err) {
			err = errdefs.System(err)
		}
	default:
		switch {
		case statusCode >= 200 && statusCode < 400:
			// it's a client error
		case statusCode >= 400 && statusCode < 500:
			err = errdefs.InvalidParameter(err)
		case statusCode >= 500 && statusCode < 600:
			err = errdefs.System(err)
		default:
			err = errdefs.Unknown(err)
		}
	}
	return err
}
