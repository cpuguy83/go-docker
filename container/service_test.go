package container

import (
	"testing"

	"github.com/cpuguy83/go-docker/testutils"
)

func newTestService(t *testing.T) *Service {
	tr, _ := testutils.NewDefaultTestTransport(t)
	return NewService(tr)
}
