package container

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

// StdType is the type of standard stream
// a writer can multiplex to.
type stdType byte

const (
	// stdinPrefix represents standard input stream type.
	stdinPrefix stdType = iota
	// stdoutPrefix represents standard output stream type.
	stdoutPrefix
	// stderrPrefix represents standard error steam type.
	stderrPrefix
	// systemerr represents errors originating from the system that make it
	// into the multiplexed stream.
	systemerr

	stdWriterPrefixLen = 8
	stdWriterFdIndex   = 0
	stdWriterSizeIndex = 4

	startingBufLen = 32*1024 + stdWriterPrefixLen + 1
)

var bufPool = &sync.Pool{New: func() interface{} { return make([]byte, startingBufLen) }}

func getStdBuff() []byte {
	return bufPool.Get().([]byte)
}

func putStdBuff(buf []byte) {
	for i := 0; i < len(buf); i++ {
		buf[i] = 0
	}
	bufPool.Put(buf)
}

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
//
// This is copied from github.com/docker/docker/pkg/stdcopy
func stdCopy(dstout, dsterr io.Writer, src io.Reader) (written int64, err error) {
	var (
		buf       = getStdBuff()
		bufLen    = len(buf)
		nr, nw    int
		er, ew    error
		out       io.Writer
		frameSize int
	)

	defer putStdBuff(buf)

	for {
		// Make sure we have at least a full header
		for nr < stdWriterPrefixLen {
			var nr2 int
			nr2, er = src.Read(buf[nr:])
			nr += nr2
			if er == io.EOF {
				if nr < stdWriterPrefixLen {
					return written, nil
				}
				break
			}
			if er != nil {
				return 0, er
			}
		}

		stream := stdType(buf[stdWriterFdIndex])
		// Check the first byte to know where to write
		switch stream {
		case stdinPrefix:
			fallthrough
		case stdoutPrefix:
			// Write on stdoutPrefix
			out = dstout
		case stderrPrefix:
			// Write on stderrPrefix
			out = dsterr
		case systemerr:
			// If we're on Systemerr, we won't write anywhere.
			// NB: if this code changes later, make sure you don't try to write
			// to outstream if Systemerr is the stream
			out = nil
		default:
			return 0, fmt.Errorf("unrecognized input header: %d", buf[stdWriterFdIndex])
		}

		// Retrieve the size of the frame
		frameSize = int(binary.BigEndian.Uint32(buf[stdWriterSizeIndex : stdWriterSizeIndex+4]))

		// Check if the buffer is big enough to read the frame.
		// Extend it if necessary.
		if frameSize+stdWriterPrefixLen > bufLen {
			buf = append(buf, make([]byte, frameSize+stdWriterPrefixLen-bufLen+1)...)
			bufLen = len(buf)
		}

		// While the amount of bytes read is less than the size of the frame + header, we keep reading
		for nr < frameSize+stdWriterPrefixLen {
			var nr2 int
			nr2, er = src.Read(buf[nr:])
			nr += nr2
			if er == io.EOF {
				if nr < frameSize+stdWriterPrefixLen {
					return written, nil
				}
				break
			}
			if er != nil {
				return 0, er
			}
		}

		// we might have an error from the source mixed up in our multiplexed
		// stream. if we do, return it.
		if stream == systemerr {
			return written, fmt.Errorf("error from daemon in stream: %s", string(buf[stdWriterPrefixLen:frameSize+stdWriterPrefixLen]))
		}

		// Write the retrieved frame (without header)
		nw, ew = out.Write(buf[stdWriterPrefixLen : frameSize+stdWriterPrefixLen])
		if ew != nil {
			return 0, ew
		}

		// If the frame has not been fully written: error
		if nw != frameSize {
			return 0, io.ErrShortWrite
		}
		written += int64(nw)

		// Move the rest of the buffer to the beginning
		copy(buf, buf[frameSize+stdWriterPrefixLen:])
		// Move the index
		nr -= frameSize + stdWriterPrefixLen
	}
}
