// +build !windows

package transport

func DefaultTransport() (*Transport, error) {
	return UnixSocketTransport("/var/run/docker.sock")
}

