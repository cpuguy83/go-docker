package container

import (
	"context"
	"sync"
	"testing"

	"github.com/cpuguy83/go-docker/system"
	"github.com/cpuguy83/go-docker/testutils"
	"github.com/cpuguy83/go-docker/transport"
	"github.com/cpuguy83/go-docker/version"
)

var (
	versionOnce          sync.Once
	negotiatedAPIVersion = ""
)

func negoiateTestAPIVersion(t testing.TB, tr transport.Doer) {
	versionOnce.Do(func() {
		ctx := context.Background()
		ctx, err := system.NewService(tr).NegotiateAPIVersion(ctx)
		if err != nil {
			t.Fatalf("error negotiating api version: %v", err)
		}
		negotiatedAPIVersion = version.APIVersion(ctx)
	})
}

func newTestService(t *testing.T, ctx context.Context) (*Service, context.Context) {
	tr, _ := testutils.NewDefaultTestTransport(t)
	if version.APIVersion(ctx) == "" {
		negoiateTestAPIVersion(t, tr)
		ctx = version.WithAPIVersion(ctx, negotiatedAPIVersion)
	}
	return NewService(tr), ctx
}
