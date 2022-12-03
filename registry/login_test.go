package registry

import (
	"context"
	"testing"

	"github.com/cpuguy83/go-docker/errdefs"
	"gotest.tools/v3/assert"
)

func TestLogin(t *testing.T) {
	getCreds := func(cfg *LoginConfig) error {
		cfg.IdentityToken = "asdf"
		return nil
	}

	s := newTestService(t)

	token, err := s.Login(context.Background(), getCreds)
	assert.ErrorIs(t, err, errdefs.ErrUnauthorized)
	assert.Assert(t, token == "")

	// TODO: Add a test for a successful login
}
