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
	// ConsumeProgress is called after a pull response is received to consume the progress messages from the response body.
	// ConSumeProgress should not return until EOF is reached on the passed in stream or it may cause the pull to be cancelled.
	// If this is not set, progress messages are discarded.
	ConsumeProgress StreamConsumer
}

// LoadOption is used as functional arguments to list images. LoadOption
// configure a LoadConfig.
type LoadOption func(config *LoadConfig) error

// Load loads container images.
func (s *Service) Load(ctx context.Context, tar io.ReadCloser, opts ...LoadOption) error {
	cfg := LoadConfig{}
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return err
		}
	}

	withListConfig := func(req *http.Request) error {
		quiet := cfg.ConsumeProgress == nil

		q := req.URL.Query()
		q.Add("quiet", strconv.FormatBool(quiet))
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

	if cfg.ConsumeProgress != nil {
		if err := cfg.ConsumeProgress(ctx, resp.Body); err != nil {
			return fmt.Errorf("consuming progress: %w", err)
		}
	} else {
		_, err := io.Copy(io.Discard, resp.Body)
		if err != nil {
			return fmt.Errorf("discarding progress: %w", err)
		}
	}

	return nil
}
