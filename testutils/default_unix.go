package testutils

import (
	"testing"

	"github.com/cpuguy83/go-docker/transport"
)

// NewDefaultTestTransport creates a default test transport
func NewDefaultTestTransport(t *testing.T, noTap bool) (*Transport, error) {
	tr, err := transport.DefaultTransport()
	if err != nil {
		return nil, err
	}

	return NewTransport(t, tr, noTap), nil
}
