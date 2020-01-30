package testutils

import (
	"testing"

	"github.com/cpuguy83/go-docker/transport"
)

// NewDefaultTestTransport creates a default test transport
func NewDefaultTestTransport(t *testing.T) *Transport {
	return NewTransport(t, transport.DefaultUnixTransport())
}
