package image_test

import (
	"testing"

	"github.com/cpuguy83/go-docker/image"
	"github.com/cpuguy83/go-docker/testutils"
)

func newTestService(t *testing.T) *image.Service {
	tr, _ := testutils.NewDefaultTestTransport(t)
	return image.NewService(tr)
}
