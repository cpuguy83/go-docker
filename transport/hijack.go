package transport

import (
	"bufio"
	"io"
	"net"
)

func newHijackedConn(conn net.Conn, buf *bufio.Reader) io.ReadWriteCloser {
	if buf.Buffered() == 0 {
		buf.Reset(nil)
		return conn
	}

	hc := &hijackConn{Conn: conn, buf: buf}

	if _, ok := conn.(closeWriter); ok {
		return &hijackConnCloseWrite{hc}
	}

	return hc
}

type hijackConn struct {
	net.Conn
	buf io.Reader
}

func (c *hijackConn) Read(p []byte) (int, error) {
	return c.buf.Read(p)
}

type hijackConnCloseWrite struct {
	*hijackConn
}

func (c *hijackConnCloseWrite) CloseWrite() error {
	return c.Conn.(closeWriter).CloseWrite()
}
