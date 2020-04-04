package system

import "github.com/cpuguy83/go-docker/transport"

// Service facilitates all communication with Docker's container endpoints.
// Create one with `NewService`
type Service struct {
	tr transport.Doer
}

// NewService creates a Service.
// This is the entrypoint to this package.
func NewService(tr transport.Doer) *Service {
	return &Service{tr}
}
