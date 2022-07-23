package streamutil

import (
	"fmt"
	"io"

	"github.com/cpuguy83/go-docker/errdefs"
)

// StdCopy will de-multiplex `src`, assuming that it contains two streams,
// previously multiplexed together using a StdWriter instance.
// As it reads from `src`, StdCopy will write to `dstout` and `dsterr`.
//
// StdCopy will read until it hits EOF on `src`. It will then return a nil error.
// In other words: if `err` is non nil, it indicates a real underlying error.
//
// `written` will hold the total number of bytes written to `dstout` and `dsterr`.
func StdCopy(dstout, dsterr io.Writer, src io.Reader) (written int64, retErr error) {
	rdr := NewStdReader(src)
	buf := make([]byte, 32*2014)

	for {
		hdr, err := rdr.Next()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return written, err
		}

		var out io.Writer

		switch hdr.Descriptor {
		case Stdout:
			out = dstout
		case Stderr:
			out = dsterr
		case Systemerr:
			// Limit the size of this message to the size of our buffer to prevent memory exhaustion
			n, err := rdr.Read(buf)
			if err != nil {
				return written, errdefs.Wrapf(err, "error while copying system error from stdio stream, truncated message=%q", buf[:n])
			}
			return written, fmt.Errorf("%s", buf[:n])
		default:
			return written, fmt.Errorf("got data for unknown stream id: %d", hdr.Descriptor)
		}

		n, err := io.CopyBuffer(out, rdr, buf)
		written += n
		if err != nil {
			return written, errdefs.Wrap(err, "got error while copying to stream")
		}
	}
}
