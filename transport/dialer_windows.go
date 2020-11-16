// +build windows

package transport

import (
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

func winDailer(path string, timeout *time.Duration) (net.Conn, error) {
	return winio.DialPipe(path, timeout)
}
