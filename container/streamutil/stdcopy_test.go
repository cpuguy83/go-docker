package streamutil

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
	"time"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

func TestStdCopyNormal(t *testing.T) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	r, w := io.Pipe()
	defer r.Close()

	var (
		nw     int64
		copied int64
	)

	data1 := []byte("hello stdout!")
	data2 := []byte("what's up stderr!")
	data3 := []byte("here's some more for you stdout")

	testCh := make(chan func(t *testing.T), 1)
	go func() {
		t.Helper()
		var err error
		copied, err = StdCopy(stdout, stderr, r)
		testCh <- func(t *testing.T) {
			assert.NilError(t, err)
			assert.Check(t, cmp.Equal(copied, nw))
			assert.Check(t, cmp.Equal(stdout.String(), string(data1)+string(data3)))
			assert.Check(t, cmp.Equal(stderr.String(), string(data2)))
		}
	}()

	stdoutHeader := [stdHeaderPrefixLen]byte{stdHeaderFdIndex: Stdout}
	stderrHeader := [stdHeaderPrefixLen]byte{stdHeaderFdIndex: Stderr}

	binary.BigEndian.PutUint32(stdoutHeader[stdHeaderSizeIndex:], uint32(len(data1)))
	_, err := w.Write(append(stdoutHeader[:], data1...))
	assert.NilError(t, err)
	nw += int64(len(data1))

	select {
	case f := <-testCh:
		f(t)
	default:
	}

	binary.BigEndian.PutUint32(stderrHeader[stdHeaderSizeIndex:], uint32(len(data2)))
	_, err = w.Write(append(stderrHeader[:], data2...))
	assert.NilError(t, err)
	nw += int64(len(data2))

	select {
	case f := <-testCh:
		f(t)
	default:
	}

	binary.BigEndian.PutUint32(stdoutHeader[stdHeaderSizeIndex:], uint32(len(data3)))
	_, err = w.Write(append(stdoutHeader[:], data3...))
	assert.NilError(t, err)
	nw += int64(len(data3))

	w.Close()

	deadline(t, 10*time.Second, testCh)
}

func TestStdCopyWithSystemErr(t *testing.T) {
	out := bytes.NewBuffer(nil)
	buf := bytes.NewBuffer(nil)
	stdoutHeader := [stdHeaderPrefixLen]byte{stdHeaderFdIndex: Stdout}
	systemErrHeader := [stdHeaderPrefixLen]byte{stdHeaderFdIndex: Systemerr}

	data1 := []byte("hello world!")
	binary.BigEndian.PutUint32(stdoutHeader[stdHeaderSizeIndex:], uint32(len(data1)))
	_, err := buf.Write(append(stdoutHeader[:], data1...))
	assert.NilError(t, err)

	badError := []byte("something really really really bad has happened")
	binary.BigEndian.PutUint32(systemErrHeader[stdHeaderSizeIndex:], uint32(len(badError)))
	_, err = buf.Write(append(systemErrHeader[:], badError...))
	assert.NilError(t, err)

	testCh := make(chan func(t *testing.T), 1)
	go func() {
		copied, err := StdCopy(out, out, buf)
		testCh <- func(t *testing.T) {
			assert.Check(t, cmp.Error(err, string(badError)))
			assert.Check(t, cmp.Equal(out.String(), string(data1)))
			assert.Check(t, cmp.Equal(copied, int64(len(data1))))
		}
	}()

	deadline(t, 10*time.Second, testCh)
}

func deadline(t *testing.T, dur time.Duration, fChan <-chan func(t *testing.T)) {
	t.Helper()

	timer := time.NewTimer(dur)
	defer timer.Stop()

	select {
	case <-timer.C:
		t.Fatal("timeout waiting")
	case f := <-fChan:
		f(t)
	}
}
