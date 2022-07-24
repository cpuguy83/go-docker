package image

import (
	"testing"

	"github.com/cpuguy83/go-docker/testutils"
	"gotest.tools/v3/assert"
)

func newTestService(t *testing.T) *Service {
	tr, err := testutils.NewDefaultTestTransport(t)
	assert.NilError(t, err)
	return NewService(tr)
}
