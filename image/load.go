package image

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
)

// LoadConfig holds the options for loading images.
type LoadConfig struct {
	Quiet bool
}

// LoadOption is used as functional arguments to list images. LoadOption
// configure a LoadConfig.
type LoadOption func(config *LoadConfig)

// Load loads container images.
func (s *Service) Load(ctx context.Context, tar io.ReadCloser, opts ...LoadOption) error {
	cfg := LoadConfig{}
	for _, o := range opts {
		o(&cfg)
	}

	withListConfig := func(req *http.Request) error {
		q := req.URL.Query()
		q.Add("quiet", strconv.FormatBool(cfg.Quiet))
		req.URL.RawQuery = q.Encode()
		req.Body = tar
		return nil
	}

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, http.MethodPost, version.Join(ctx, "/images/load"), withListConfig)
	})
	if err != nil {
		return fmt.Errorf("loading images: %w", err)
	}
	defer resp.Body.Close()
	return nil
}
