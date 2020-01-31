package system

import "github.com/cpuguy83/go-docker/transport"

type Service struct {
	tr transport.Doer
}

func NewService(tr transport.Doer) *Service {
	return &Service{tr}
}
