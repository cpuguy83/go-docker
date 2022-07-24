package image

import "github.com/cpuguy83/go-docker/transport"

// Service facilitates all communication with Docker's container endpoints.
// Create one with `NewService`
type Service struct {
	tr transport.Doer
}

// NewService creates a new Service.
func NewService(tr transport.Doer) *Service {
	return &Service{tr: tr}
}
