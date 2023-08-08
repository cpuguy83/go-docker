package testutils

import (
	"os"
	"testing"

	"github.com/cpuguy83/go-docker/testutils/assert"
	"github.com/cpuguy83/go-docker/transport"
)

// NewDefaultTestTransport creates a default test transport
func NewDefaultTestTransport(t *testing.T, noTap bool) (*Transport, error) {
	if os.Getenv("DOCKER_HOST") != "" || os.Getenv("DOCKER_CONTEXT") != "" {
		t.Log("Using docker cli transport")
		tr := transport.FromDockerCLI(func(cfg *transport.DockerCLIConnectionConfig) error {
			cfg.StderrPipe = &testWriter{t}
			return nil
		})
		return NewTransport(t, tr, noTap), nil
	}

	t.Log("Using system default transport")
	tr, err := transport.DefaultTransport()
	assert.NilError(t, err)

	return NewTransport(t, tr, noTap), nil
}

type testWriter struct {
	t *testing.T
}

func (t *testWriter) Write(p []byte) (int, error) {
	t.t.Helper()
	t.t.Log(string(p))
	return len(p), nil
}
