package image_test

import (
	"testing"

	"github.com/cpuguy83/go-docker/image"
	"github.com/cpuguy83/go-docker/testutils"
	"github.com/cpuguy83/go-docker/transport"
	"gotest.tools/assert"
)

func newTestService(t *testing.T) *image.Service {
	tr, err := testutils.NewDefaultTestTransport(t)
	assert.NilError(t, err)
	return image.NewService(tr)
}

func newTestServiceNormalTransport(t *testing.T) *image.Service {
	tr, err := transport.DefaultTransport()
	assert.NilError(t, err)
	return image.NewService(tr)
}
