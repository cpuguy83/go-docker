package streamutil

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/cpuguy83/go-docker/errdefs"

	"gotest.tools/v3/assert"
)

func TestStreamReader(t *testing.T) {
	stream := bytes.NewBuffer(nil)

	line1 := []byte("this is line 1")
	line2 := []byte("this is line 2")

	header := [stdHeaderPrefixLen]byte{stdHeaderFdIndex: Stdout}
	binary.BigEndian.PutUint32(header[stdHeaderSizeIndex:], uint32(len(line1)))
	stream.Write(header[:])
	stream.Write(line1)

	header = [stdHeaderPrefixLen]byte{stdHeaderFdIndex: Stderr}
	binary.BigEndian.PutUint32(header[stdHeaderSizeIndex:], uint32(len(line2)))
	stream.Write(header[:])
	stream.Write(line2)

	rdr := NewStdReader(stream)

	buf := make([]byte, len(line1))

	n, err := rdr.Read(buf)
	assert.ErrorType(t, err, io.EOF)
	assert.Equal(t, n, 0)

	h, err := rdr.Next()
	assert.NilError(t, err)
	assert.Equal(t, h.Descriptor, Stdout)
	assert.Equal(t, h.Size, len(line1))

	_, err = rdr.Next()
	assert.Assert(t, errdefs.IsConflict(err), err)

	n, err = rdr.Read(buf)
	assert.NilError(t, err)
	assert.Equal(t, n, len(line1))

	n, err = rdr.Read(buf)
	assert.ErrorType(t, err, io.EOF)
	assert.Equal(t, n, 0)

	h, err = rdr.Next()
	assert.NilError(t, err)
	assert.Equal(t, h.Descriptor, Stderr)
	assert.Equal(t, h.Size, len(line2))
	assert.Equal(t, string(buf), string(line1))

	buf = make([]byte, len(line2))
	n, err = rdr.Read(buf)
	assert.NilError(t, err)
	assert.Equal(t, n, len(line2))
	assert.Equal(t, string(buf), string(line2))

	_, err = rdr.Next()
	assert.ErrorType(t, err, io.EOF)
}
