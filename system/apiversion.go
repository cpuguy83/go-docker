package system

import (
	"context"

	"github.com/cpuguy83/go-docker/version"
)

// NegoitateAPIVersion negotiates the API version to use with the server.
// The returned context stores the version.
// Pass that ctx into calls that you want to use this negoiated version with.
func (s *Service) NegotiateAPIVersion(ctx context.Context) (context.Context, error) {
	p, err := s.Ping(ctx)
	if err != nil {
		if p.APIVersion == "" {
			return ctx, err
		}
	}
	return version.Negotiate(ctx, p.APIVersion), nil
}
