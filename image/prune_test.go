package image

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	"github.com/cpuguy83/go-docker/image/imageapi"
	"gotest.tools/v3/assert"
)

func TestPrune(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	buildContainers := func(t *testing.T) {
		buildContainer(t, "test-image:positive", "test-image", "positive=true")
		buildContainer(t, "test-image:negative", "test-image", "positive=false")
		// Add a dummy image that should never be discovered in the tests below.
		// In particualar, this tests that the interaction between Label and
		// NotLabel filter works as intended.
		buildContainer(t, "other-image:all", "other-image")
	}
	cleanup := func(t *testing.T) {
		_, err := s.Prune(ctx, func(config *PruneConfig) {
			config.Filters.Dangling = []string{"false"}
			config.Filters.Label = []string{"test-image"}
		})
		assert.NilError(t, err)
		_, err = s.Prune(ctx, func(config *PruneConfig) {
			config.Filters.Dangling = []string{"false"}
			config.Filters.Label = []string{"other-image"}
		})
		assert.NilError(t, err)
	}

	extract := func(prune imageapi.Prune) []string {
		var l []string
		for _, deleted := range prune.ImagesDeleted {
			if deleted.Untagged != "" {
				l = append(l, deleted.Untagged)
			}
		}
		sort.Strings(l)
		return l
	}

	t.Run("all", func(t *testing.T) {
		buildContainers(t)
		defer cleanup(t)

		rep, err := s.Prune(ctx, func(config *PruneConfig) {
			config.Filters.Dangling = []string{"false"}
			config.Filters.Label = []string{"test-image"}
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, extract(rep), []string{"test-image:negative", "test-image:positive"})
	})

	t.Run("positive", func(t *testing.T) {
		buildContainers(t)
		defer cleanup(t)

		rep, err := s.Prune(ctx, func(config *PruneConfig) {
			config.Filters.Dangling = []string{"false"}
			config.Filters.Label = []string{"positive=true"}
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, extract(rep), []string{"test-image:positive"})
	})

	t.Run("non-positive", func(t *testing.T) {
		buildContainers(t)
		defer cleanup(t)

		rep, err := s.Prune(ctx, func(config *PruneConfig) {
			config.Filters.Dangling = []string{"false"}
			config.Filters.Label = []string{"test-image"}
			config.Filters.NotLabel = []string{"positive=true"}
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, extract(rep), []string{"test-image:negative"})
	})
}

// buildContainer builds the container by executing the docker CLI until we have
// extended the image service to provide the build endpoint.
func buildContainer(t *testing.T, tag string, labels ...string) {
	dockerfile := `FROM scratch
CMD ["hello"]
`
	dir := t.TempDir()
	assert.NilError(t, os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(dockerfile), 0666))

	args := []string{"build", dir, "--tag=" + tag}
	for _, label := range labels {
		args = append(args, "--label="+label)
	}

	rep, err := exec.Command("docker", args...).CombinedOutput()
	assert.NilError(t, err, string(rep))
}
