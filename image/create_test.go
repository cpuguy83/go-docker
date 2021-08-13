package image_test

import (
	"context"
	"fmt"
	"os/exec"
	"testing"

	"github.com/cpuguy83/go-docker/image"
	"gotest.tools/assert"
)

func TestCreate(t *testing.T) {
	ctx := context.Background()
	for i, s := range []*image.Service{newTestService(t), newTestServiceNormalTransport(t)} {
		t.Run(fmt.Sprintf("transport %d", i), func(t *testing.T) {
			err := s.Create(ctx)
			assert.Assert(t, err != nil, "expected create with no options to fail")

			cmd := exec.Command("docker", "image", "rm", "-f", "busybox")
			assert.NilError(t, cmd.Run(), "expected running cleanup command to succeed")

			err = s.Create(ctx, func(config *image.CreateConfig) {
				config.FromImage = "busybox"
				config.Tag = "latest"
				config.Repo = "dockerhub.io"
			})
			assert.NilError(t, err, "expected pulling busybox to succeed")
			images, err := s.List(ctx, func(config *image.ListConfig) {
				config.Filter.Reference = append(config.Filter.Reference, "busybox:latest")
			})
			assert.NilError(t, err, "expected listing images to succeed")
			assert.Assert(t, len(images) == 1, "expected image to be found")
		})
	}
}
