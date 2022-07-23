package streamutil

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/cpuguy83/go-docker/errdefs"
)

const (
	// Stdin represents standard input stream type.
	Stdin = iota
	// Stdout represents standard output stream type.
	Stdout
	// Stderr represents standard error steam type.
	Stderr
	// Systemerr represents errors originating from the system that make it
	// into the multiplexed stream.
	Systemerr

	stdHeaderPrefixLen = 8
	stdHeaderFdIndex   = 0
	stdHeaderSizeIndex = 4
)

type StdReader struct {
	// The whole stream
	rdr io.Reader
	// The current frame only
	curr   io.Reader
	currNR int

	buf []byte
	hdr *StdHeader
	err error
}

// StdHeader is a descriptor for a stdio stream frame
// It gets used in conjunction with StdReader
type StdHeader struct {
	Descriptor int
	Size       int
}

// NewStdReader creates a reader for consuming a stdio stream.
// Specifically this can be used for processing a streaming following Docker's stdio stream format where the stream
// has an 8 byte header describing the message including what stdio stream (stdout, stderr) it belongs to and the size
// of the message followed by the message itself.
func NewStdReader(rdr io.Reader) *StdReader {
	return &StdReader{rdr: rdr, buf: make([]byte, stdHeaderPrefixLen)}
}

// Next returns the next stream header
//
// If Next is called before consuming the previous descriptor, an errdefs.Conflict error will be returned (check with errdefs.IsConflict(err)).
// If there are any other errors processing the stream the error will be returned immediately, and again on all subsequent calls.
func (s *StdReader) Next() (*StdHeader, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.curr != nil {
		// Cannot proceed until all data is drained for s.curr
		return nil, errdefs.Conflict("unconsumed data in stream; read to EOF before calling again")
	}

	_, err := io.ReadFull(s.rdr, s.buf)
	if err != nil {
		if err == io.EOF {
			s.err = err
		} else {
			s.err = errdefs.Wrap(err, "error reading log message header")
		}
		return nil, s.err
	}

	if s.hdr == nil {
		s.hdr = &StdHeader{}
	}

	fd := int(s.buf[stdHeaderFdIndex])
	switch fd {
	case Stdin:
		fallthrough
	case Stdout, Stderr, Systemerr:
		s.hdr.Descriptor = fd
	default:
		s.err = fmt.Errorf("malformed stream, got unexpected stream descriptor in header %d", fd)
		return nil, s.err
	}

	s.hdr.Size = int(binary.BigEndian.Uint32(s.buf[stdHeaderSizeIndex : stdHeaderSizeIndex+4]))
	s.curr = io.LimitReader(s.rdr, int64(s.hdr.Size))

	return s.hdr, nil
}

// Read reads up to p bytes from the stream.
// You must read until EOF before calling Next again.
func (s *StdReader) Read(p []byte) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	if s.curr == nil {
		return 0, io.EOF
	}

	n, err := s.curr.Read(p)
	if err == io.EOF {
		s.curr = nil
	}
	s.currNR += n
	if s.currNR == s.hdr.Size {
		s.currNR = 0
		s.curr = nil
	}

	return n, err
}
