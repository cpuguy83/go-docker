// +build linux

package transport

import (
	"errors"
	"net"
	"time"
)

func winDailer(path string, timeout *time.Duration) (net.Conn, error) {
	return nil, errors.New("windows dailers are not supported on unix platforms")
}

