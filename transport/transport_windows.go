// +build windows

package transport

func DefaultTransport() (*Transport, error) {
	return NpipeTransport("//./pipe/docker_engine")
}
