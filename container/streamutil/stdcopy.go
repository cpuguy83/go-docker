package streamutil

import (
	"io"
	"strings"

	"github.com/pkg/errors"
)

// StdCopy will de-multiplex `src`, assuming that it contains two streams,
// previously multiplexed together using a StdWriter instance.
// As it reads from `src`, StdCopy will write to `dstout` and `dsterr`.
//
// StdCopy will read until it hits EOF on `src`. It will then return a nil error.
// In other words: if `err` is non nil, it indicates a real underlying error.
//
// `written` will hold the total number of bytes written to `dstout` and `dsterr`.
func StdCopy(dstout, dsterr io.Writer, src io.Reader) (written int64, err error) {
	rdr := NewStdReader(src)
	buf := make([]byte, 32*2014)

	for {
		hdr, err := rdr.Next()
		if err != nil {
			return written, err
		}

		var out io.Writer

		switch hdr.Descriptor {
		case Stdout:
			out = dstout
		case Stderr:
			out = dsterr
		case Systemerr:
			sb := &strings.Builder{}
			// Limit the size of this message to the size of our buffer to prevent memory exhaustion
			if _, err := io.CopyBuffer(sb, io.LimitReader(rdr, int64(len(buf))), buf); err != nil {
				sb.Reset()
				return written, errors.Wrapf(err, "error while copying system error from stdio stream, truncated message=%q", sb)
			}
			sb.Reset()
			return written, errors.Errorf("%s", sb)
		default:
			return written, errors.Errorf("got data for unknown stream id: %d", hdr.Descriptor)
		}

		n, err := io.CopyBuffer(out, rdr, buf)
		written += n
		if err != nil {
			return written, errors.Wrap(err, "got error while copying to stream")
		}
	}
}
