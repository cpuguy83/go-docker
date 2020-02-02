package container

import (
	"errors"
	"io"

	"github.com/cpuguy83/go-docker/container/streamutil"
)

// StdCopy is a modified version of io.Copy.
//
// StdCopy will demultiplex `src`, assuming that it contains two streams,
// previously multiplexed together using a StdWriter instance.
// As it reads from `src`, StdCopy will write to `dstout` and `dsterr`.
//
// StdCopy will read until it hits EOF on `src`. It will then return a nil error.
// In other words: if `err` is non nil, it indicates a real underlying error.
//
// `written` will hold the total number of bytes written to `dstout` and `dsterr`.
func stdCopy(dstout, dsterr io.Writer, src io.Reader) (written int64, err error) {
	rdr := streamutil.NewStdReader(src)
	buf := make([]byte, 32*2014)

	for {
		hdr, err := rdr.Next()
		if err != nil {
			return written, err
		}

		for nr := 0; nr < hdr.Size; {
			n, err := rdr.Read(buf)
			if err != nil && err != io.EOF {
				return written, err
			}
			nr += n

			switch hdr.Descriptor {
			case streamutil.Stdout:
				w, err := dstout.Write(buf[:n])
				written += int64(w)
				if err != nil {
					return written, err
				}
			case streamutil.Stderr:
				w, err := dsterr.Write(buf[:n])
				written += int64(w)
				if err != nil {
					return written, err
				}
			case streamutil.Systemerr:
				// TODO: return an error with the message set to the read data?
			default:
				return written, errors.New("got data for unknown stream")
			}
			if err == io.EOF {
				break
			}
		}
	}
}
