package container

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cpuguy83/go-docker/container/containerapi"
	"github.com/cpuguy83/go-docker/transport"
	"github.com/pkg/errors"
)

// DefaultInspectDecodeLimitBytes is the default value used for limit how much data is read from the inspect response.
const DefaultInspectDecodeLimitBytes = 64 * 1024

// InspectConfig holds the options for inspecting a container
type InspectConfig struct {
	// Only read `DecodeLimitBytes` bytes from the inspect response
	// Set to -1 for unlimited.
	DecodeLimitBytes int64
	// Allows callers of `Inspect` to unmarshal to any object rather than only the built-in types.
	// This is useful for anyone wrapping the API and providing more metadata (e.g. classic swarm)
	// To must be a pointer or it may cause a panic.
	// If `To` is provided, `Inspect`'s returned container object may be empty.
	To interface{}
}

// InspectOption is used as functional arguments to inspect a container
// InspectOptions configure an InspectConfig.
type InspectOption func(config *InspectConfig)

// Inspect fetches detailed information about a container.
func (s *Service) Inspect(ctx context.Context, name string, opts ...InspectOption) (containerapi.ContainerInspect, error) {
	return handleInspect(ctx, s.tr, name, opts...)
}

func handleInspect(ctx context.Context, tr transport.Doer, name string, opts ...InspectOption) (containerapi.ContainerInspect, error) {
	cfg := InspectConfig{
		DecodeLimitBytes: DefaultInspectDecodeLimitBytes,
	}
	for _, o := range opts {
		o(&cfg)
	}

	var c containerapi.ContainerInspect

	resp, err := tr.Do(ctx, http.MethodGet, "/containers/"+name+"/json")
	if err != nil {
		return c, err
	}

	defer resp.Body.Close()

	var rdr io.Reader = resp.Body

	if cfg.DecodeLimitBytes > 0 {
		rdr = io.LimitReader(rdr, cfg.DecodeLimitBytes)
	}

	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		return c, nil
	}

	if cfg.To != nil {
		if err := json.Unmarshal(data, cfg.To); err != nil {
			return c, errors.Wrap(err, "error unmarshalling to requested type")
		}
		return c, nil
	}

	if err := json.Unmarshal(data, &c); err != nil {
		return c, errors.Wrap(err, "error unmarshalling container json")
	}

	return c, nil
}

// Inspect fetches detailed information about the container.
func (c *Container) Inspect(ctx context.Context, opts ...InspectOption) (containerapi.ContainerInspect, error) {
	return handleInspect(ctx, c.tr, c.id, opts...)
}
